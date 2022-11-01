package gogossip

import (
	"encoding/json"
	"fmt"

	lru "github.com/hashicorp/golang-lru"
)

type Gossiper struct {
	messageCache lru.Cache
}

func (g *Gossiper) handler(buf []byte) error {
	payload, packetType, encType, err := RemoveLabelFromPacket(buf)
	if err != nil {
		return err
	}
	//
	_, _ = payload, encType
	switch packetType {
	case PullReqestType:
		// 1. Payload decryption
		//
		// 2. Unmarshal
		var request PullRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			return err
		}
		// 3. Checks requested data
		value, ok := g.messageCache.Get(request.Target)
		if !ok {
			_ = "skip"
		}
		data, ok := value.([]byte)
		if !ok {
			panic("critical error")
		}
		// 4. Build the message
		msg := new(PullResponse)
		msg.SetID(idGenerator())
		msg.Data = data
		// 5. Send to requestor
	case PullResponseType:
		msg := new(PullRequest)
		if err := json.Unmarshal(payload, &msg); err != nil {
			return err
		}
		if g.Exist(msg.ID()) {
			_ = "skip"
		}
		// 1. Select random peers
		// 2. Make PullRequest
		// 3. Encryption
		// 4. Send
	}
	return fmt.Errorf("invalid packet type %d", packetType)
}

func (g *Gossiper) Exist(id [8]byte) bool {
	return g.messageCache.Contains(id)
}
