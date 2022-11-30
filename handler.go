package gogossip

import (
	"encoding/binary"
	"encoding/json"
	"net"
)

func (g *Gossiper) handler(buf []byte, sender *net.UDPAddr) {
	label, plain, err := unmarshalWithDecryption(buf, g.cfg.Passphrase)
	if err != nil {
		g.logger.Printf("handler: unmarshalPayloadWithDecryption failure, %v (<-- %v)\n", err, sender)
		return
	}

	// Packet to use when we need to send a response.
	var packets []Packet
	defer func() {
		for _, packet := range packets {
			g.send(packet, label.encryptType)
		}
	}()

	switch label.packetType {
	case PullRequestType:
		packets = g.pullRequestHandle(plain, sender)
	case PullResponseType:
		g.pullResponseHandle(plain)
	default:
		g.logger.Printf("hander: invalid packet detectd, type: %d (<-- %v)\n", label.packetType, sender)
	}
}

// If it is split into bytes after marshaling, the entire data will be lost if lost.
// Transmit the split data into packets.
func (g *Gossiper) pullRequestHandle(payload []byte, sender *net.UDPAddr) []Packet {
	kl, vl := g.messages.items()
	if len(kl) != len(vl) {
		panic("pullRequestHandle: invalid protocol detected, different key value sizes in the packet")
	}
	if len(kl) == 0 {
		return nil
	}

	var packets []Packet

	i := 0
	for i < len(kl) {
		r := &PullResponse{
			to:     sender,
			Keys:   make([][8]byte, 0),
			Values: make([][]byte, 0),
		}

		size := 0
		for ; i < len(kl); i++ {
			r.Keys = append(r.Keys, kl[i])
			r.Values = append(r.Values, vl[i])

			size += binary.Size(kl[i]) + binary.Size(vl[i])
			if size >= actualPayloadSize {
				r.Keys = r.Keys[:len(r.Keys)-1]
				r.Values = r.Values[:len(r.Values)-1]
				i--
				break
			}
		}
		packets = append(packets, r)
	}

	return packets
}

func (g *Gossiper) pullResponseHandle(payload []byte) {
	var msg PullResponse
	if err := json.Unmarshal(payload, &msg); err != nil {
		g.logger.Printf("pullResponseHandle: Unmarshal failure, %v\n", err)
		return
	}
	if len(msg.Keys) != len(msg.Values) {
		panic("pullResponseHandle: invalid protocol detected, different key value sizes in the packet")
	}

	for i := 0; i < len(msg.Keys); i++ {
		g.push(msg.Keys[i], msg.Values[i], true)
	}
}
