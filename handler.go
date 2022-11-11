package gogossip

import (
	"encoding/json"
	"log"
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
				log.Printf("handler: encryption failure, %v", err)
				return
			}
			p := BytesToLabel([]byte{packet.Kind(), byte(encType)}).combine(cipher)
			if _, err := g.transport.WriteToUDP(p, sender); err != nil {
				log.Printf("handler: transport filaure, %v", err)
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
		log.Printf("hander: invalid packet detectd, type: %d sender: %s", label.packetType, sender.String())
	}
}

func (g *Gossiper) pullRequestHandle(payload []byte, enctype EncryptType, sender *net.UDPAddr) Packet {
	kl, vl := g.messages.itemsWithTouch(sender.String())
	if len(kl) != len(vl) {
		log.Printf("pullRequestHandle: invalid protocol detected, different key value sizes in the packet")
		return nil
	}

	// TODO(dbadoy): Send it in multiple pieces. Sometimes occur error
	// about message too big.
	return &PullResponse{atomic.AddUint32(&g.seq, 1), kl, vl}
}

func (g *Gossiper) pullResponseHandle(payload []byte, encType EncryptType) {
	var msg PullResponse
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	if len(msg.Keys) != len(msg.Values) {
		log.Printf("pullResponseHandle: invalid protocol detected, different key value sizes in the packet")
		return
	}

	for i := 0; i < len(msg.Keys); i++ {
		g.push(msg.Keys[i], msg.Values[i])
	}
}
