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
	var packets []packet
	defer func() {
		for _, p := range packets {
			g.send(p, label.encryptType)
		}
	}()

	switch label.packetType {
	case pullRequestType:
		packets = g.handlePullRequest(plain, sender)
	case pullResponseType:
		g.handlePullResponse(plain)
	default:
		g.logger.Printf("hander: invalid packet detectd, type: %d (<-- %v)\n", label.packetType, sender)
	}
}

// If it is split into bytes after marshaling, the entire data will be lost if lost.
// Transmit the split data into packets.
func (g *Gossiper) handlePullRequest(payload []byte, sender *net.UDPAddr) []packet {
	kl, vl := g.messages.items()
	if len(kl) != len(vl) {
		panic("handlePullRequest: invalid protocol detected, different key value sizes in the packet")
	}
	if len(kl) == 0 {
		return nil
	}

	var packets []packet

	i := 0
	for i < len(kl) {
		r := &pullResponse{
			to:     sender,
			Keys:   make([][8]byte, 0),
			Values: make([][]byte, 0),
		}

		accum := 0
		for ; i < len(kl); i++ {
			prop := binary.Size(kl[i]) + binary.Size(vl[i])
			if accum+prop >= actualPayloadSize {
				break
			}

			r.Keys = append(r.Keys, kl[i])
			r.Values = append(r.Values, vl[i])
			accum += prop
		}
		packets = append(packets, r)
	}

	return packets
}

func (g *Gossiper) handlePullResponse(payload []byte) {
	var msg pullResponse
	if err := json.Unmarshal(payload, &msg); err != nil {
		g.logger.Printf("handlePullResponse: Unmarshal failure, %v\n", err)
		return
	}
	if len(msg.Keys) != len(msg.Values) {
		panic("handlePullResponse: invalid protocol detected, different key value sizes in the packet")
	}

	for i := 0; i < len(msg.Keys); i++ {
		g.push(msg.Keys[i], msg.Values[i], true)
	}
}
