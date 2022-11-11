package gogossip

import (
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"
)

func (g *Gossiper) handler(buf []byte, sender *net.UDPAddr) {
	label, payload := SplitLabel(buf)
	encType := EncryptType(label.encryptType)

	plain, err := DecryptPayload(encType, g.cfg.passphrase, payload)
	if err != nil {
		return
	}

	// Packet to use when we need to send a response.
	var packet Packet
	defer func() {
		if packet != nil {
			cipher, err := EncryptPacket(encType, g.cfg.passphrase, packet)
			if err != nil {
				return
			}
			p := BytesToLabel([]byte{packet.Kind(), byte(encType)}).combine(cipher)
			if _, err := g.transport.WriteToUDP(p, sender); err != nil {
				return
			}
		}
	}()

	switch label.packetType {
	case PullRequestType:
		packet = g.pullRequestHandle(plain, encType, sender)
	case PullResponseType:
		g.pullResponseHandle(plain, encType)
	default:
		fmt.Printf("invalid packet type %d", label.packetType)
	}
}

func (g *Gossiper) pullRequestHandle(payload []byte, enctype EncryptType, sender *net.UDPAddr) Packet {
	kl, vl := g.messages.itemsWithTouch(sender.String())
	if len(kl) != len(vl) {
		panic("invalid protocol detected")
	}

	return &PullResponse{atomic.AddUint32(&g.seq, 1), kl, vl}
}

func (g *Gossiper) pullResponseHandle(payload []byte, encType EncryptType) {
	var msg PullResponse
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	if len(msg.Keys) != len(msg.Values) {
		panic("invalid protocol detected")
	}

	for i := 0; i < len(msg.Keys); i++ {
		g.push(msg.Keys[i], msg.Values[i])
	}
}
