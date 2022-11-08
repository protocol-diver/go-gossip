package gogossip

import (
	"encoding/json"
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
	case PushMessageType:
		packet = g.pushMessageHandle(plain, encType, sender)
	case PushAckType:
		g.pushAckHandle(plain, encType)
	case PullSyncType:
		packet = g.pullSyncHandle(plain, encType, sender)
	case PullRequestType:
		packet = g.pullRequestHandle(plain, encType, sender)
	case PullResponseType:
		g.pullResponseHandle(plain, encType)
	}
	// ("invalid packet type %d", packetType)
}

func (g *Gossiper) pushMessageHandle(payload []byte, encType EncryptType, sender *net.UDPAddr) Packet {
	var msg PushMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil
	}

	value := g.get(msg.Key)
	if value != nil {
		return nil
	}

	// Send gossip message to pipe if receive newly message.
	g.add(msg.Key, msg.Data)

	return &PushAck{atomic.AddUint32(&g.seq, 1), msg.Key}
}

func (g *Gossiper) pushAckHandle(payload []byte, encType EncryptType) {
	var msg PushAck
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}

	g.ackMu.Lock()
	// Ignore packets received after AckTimeout
	if g.ackChan == nil {
		g.ackMu.Unlock()
		return
	}
	ch := g.ackChan[msg.Key]
	g.ackMu.Unlock()

	ch <- msg
}

func (g *Gossiper) pullSyncHandle(payload []byte, encType EncryptType, sender *net.UDPAddr) Packet {
	var msg PullSync
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil
	}

	return &PullRequest{atomic.AddUint32(&g.seq, 1), msg.Target}
}

func (g *Gossiper) pullRequestHandle(payload []byte, enctype EncryptType, sender *net.UDPAddr) Packet {
	var msg PullRequest
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil
	}

	value := g.get(msg.Target)
	if value == nil {
		return nil
	}

	return &PullResponse{atomic.AddUint32(&g.seq, 1), msg.Target, value}
}

func (g *Gossiper) pullResponseHandle(payload []byte, encType EncryptType) {
	var msg PullResponse
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}

	value := g.get(msg.Target)
	if value != nil {
		return
	}

	g.add(msg.Target, msg.Data)
}
