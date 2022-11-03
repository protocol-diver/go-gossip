package gogossip

import (
	"encoding/json"
	"net"
	"sync/atomic"
)

func (g *Gossiper) handler(buf []byte, sender *net.UDPAddr) {
	payload, packetType, encType, err := RemoveLabelFromPacket(buf)
	if err != nil {
		return
	}
	if encType != TEMP_NONE_ENC {
		cipher := NewCipher(encType)
		payload, err = cipher.Decrypt(g.cfg.passphrase, payload)
		if err != nil {
			return
		}
	}
	switch packetType {
	case PushMessageType:
		g.pushMessageHandle(payload, encType, sender)
	case PushAckType:
		g.pushAckHandle(payload, encType)
	case PullSyncType:
		g.pullSyncHandle(payload, encType, sender)
	case PullRequestType:
		g.pullRequestHandle(payload, encType, sender)
	case PullResponseType:
		g.pullResponseHandle(payload, encType)
	}
	// ("invalid packet type %d", packetType)
}

func (g *Gossiper) pushMessageHandle(payload []byte, encType EncryptType, sender *net.UDPAddr) {
	var msg PushMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}

	value := g.get(msg.Key)
	if value != nil {
		return
	}

	g.add(msg.Key, msg.Data)

	response := PushAck{atomic.AddUint32(&g.seq, 1), msg.Key}

	buf, err := AddLabelFromPacket(&response, encType)
	if err != nil {
		return
	}
	if _, err := g.transport.WriteToUDP(buf, sender); err != nil {
		return
	}
}

func (g *Gossiper) pushAckHandle(payload []byte, encType EncryptType) {
	var msg PushAck
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}

	g.ackMu.Lock()
	ch := g.ackChan[msg.Key]
	g.ackMu.Unlock()

	ch <- &msg
}

func (g *Gossiper) pullSyncHandle(payload []byte, encType EncryptType, sender *net.UDPAddr) {
	var msg PullSync
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}

	request := PullRequest{atomic.AddUint32(&g.seq, 1), msg.Target}

	buf, err := AddLabelFromPacket(&request, encType)
	if err != nil {
		return
	}
	if _, err := g.transport.WriteToUDP(buf, sender); err != nil {
		return
	}
}

func (g *Gossiper) pullRequestHandle(payload []byte, enctype EncryptType, sender *net.UDPAddr) {
	var msg PullRequest
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}

	value := g.get(msg.Target)
	if value == nil {
		return
	}

	response := PullResponse{atomic.AddUint32(&g.seq, 1), msg.Target, value}

	buf, err := AddLabelFromPacket(&response, enctype)
	if err != nil {
		return
	}
	if _, err := g.transport.WriteToUDP(buf, sender); err != nil {
		return
	}
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
