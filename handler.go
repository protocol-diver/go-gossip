package gogossip

import (
	"encoding/binary"
	"encoding/json"
	"net"
)

func (g *Gossiper) handler(buf []byte, sender *net.UDPAddr) {
	label, payload, err := splitLabel(buf)
	if err != nil {
		g.logger.Printf("handler: splitLabel failure, %v", err)
		return
	}

	encType := EncryptType(label.encryptType)

	plain, err := DecryptPayload(payload, encType, g.cfg.Passphrase)
	if err != nil {
		g.logger.Printf("handler: DecryptPayload failure, %v", err)
		return
	}

	// Packet to use when we need to send a response.
	var packets []Packet
	defer func() {
		for _, packet := range packets {
			g.send(packet, encType)
		}
	}()

	switch label.packetType {
	case PullRequestType:
		packets = g.pullRequestHandle(plain, encType, sender)
	case PullResponseType:
		g.pullResponseHandle(plain, encType)
	default:
		g.logger.Printf("hander: invalid packet detectd, type: %d sender: %s", label.packetType, sender.String())
	}
}

// If it is split into bytes after marshaling, the entire data will be lost if lost.
// Transmit the split data into packets.
func (g *Gossiper) pullRequestHandle(payload []byte, enctype EncryptType, sender *net.UDPAddr) []Packet {
	kl, vl := g.messages.itemsWithTouch(sender.String())
	if len(kl) != len(vl) {
		panic("pullRequestHandle: invalid protocol detected, different key value sizes in the packet")
	}

	var packets []Packet

	i := 0
	for i < len(kl) {
		prealloc := actualDataSize / (8 + binary.Size(vl[i]))
		r := &PullResponse{
			to:     sender,
			Keys:   make([][8]byte, 0, prealloc),
			Values: make([][]byte, 0, prealloc),
		}

		size := 0
		for ; i < len(kl); i++ {
			r.Keys = append(r.Keys, kl[i])
			r.Values = append(r.Values, vl[i])

			size += binary.Size(kl[i]) + binary.Size(vl[i])
			if size >= actualDataSize {
				break
			}
		}
		packets = append(packets, r)
	}

	return packets
}

func (g *Gossiper) pullResponseHandle(payload []byte, encType EncryptType) {
	var msg PullResponse
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	if len(msg.Keys) != len(msg.Values) {
		panic("pullResponseHandle: invalid protocol detected, different key value sizes in the packet")
	}

	for i := 0; i < len(msg.Keys); i++ {
		g.push(msg.Keys[i], msg.Values[i], true)
	}
}
