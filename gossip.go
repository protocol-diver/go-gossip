package gogossip

import (
	"encoding/json"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type Gossiper struct {
	latest map[string]time.Time

	messagePipe  chan []byte
	messageCache lru.Cache
}

func (g *Gossiper) Messages() chan []byte {
	return g.messagePipe
}

func (g *Gossiper) handler(buf []byte) error {
	payload, packetType, encType, err := RemoveLabelFromPacket(buf)
	if err != nil {
		return err
	}
	_ = encType

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
		data := g.get(request.Target)
		if data == nil {
			_ = "skip"
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
		value := g.get(msg.ID())
		if value != nil {
			_ = "skip"
		}
		// temp
		if true {
			g.messageCache.Add(msg.ID(), []byte{})
			g.messagePipe <- []byte{}
		}
		_ = value
		// 1. Select random peers
		// 2. Make PullRequest
		// 3. Encryption
		// 4. Send
	}
	return fmt.Errorf("invalid packet type %d", packetType)
}

func (g *Gossiper) get(id [8]byte) []byte {
	v, ok := g.messageCache.Get(id)
	if !ok {
		return nil
	}
	return v.([]byte)
}
