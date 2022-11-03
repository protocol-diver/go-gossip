package gogossip

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"
)

const (
	// cfg
	GossipNumber = 2

	// Set to twice the predicted value of messages to occur per second
	MessageCacheSize = 0
)

type Gossiper struct {
	discovery Discovery

	messagePipe  chan []byte
	messageCache lru.Cache

	cfg *Config
}

func (g *Gossiper) Push(data []byte) error {
	return nil
}

func (g *Gossiper) Messages() chan []byte {
	return g.messagePipe
}

func (g *Gossiper) SelectRandomPeers() []string {
	return nil
}

func (g *Gossiper) handler(buf []byte) error {
	payload, packetType, encType, err := RemoveLabelFromPacket(buf)
	if err != nil {
		return err
	}
	_, _ = payload, encType

	switch packetType {
	case PushMessageType:
		// 1. Store / Ignore if already exist
		// 2. Store? -> Send PushAckType to sender
	case PushAckType:
		// 1. Counts number of ACK
		// 2. Done or multicast PullSyncType
	case PullSyncType:
		// 1. Send PullRequest to sender
	case PullRequestType:
		// 1. Find requested data
		// 2. Marshal and encryption
		// 3. Make PullResponse
		// 4. Send to requestor
	case PullResponseType:
		// 1. Store / Ignore if already exist
	}
	return fmt.Errorf("invalid packet type %d", packetType)
}

func (g *Gossiper) get(id uint) []byte {
	v, ok := g.messageCache.Get(id)
	if !ok {
		return nil
	}
	return v.([]byte)
}
