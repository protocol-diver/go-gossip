package gogossip

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	// cfg
	GossipNumber = 2

	// Set to twice the predicted value of messages to occur per second
	MessageCacheSize = 0
)

type Gossiper struct {
	seq uint32

	discovery Discovery
	transport Transport

	messagePipe chan []byte

	ackChan map[[8]byte]chan Packet
	ackMu   sync.Mutex

	messageCache lru.Cache

	cfg *Config
}

func (g *Gossiper) dispatch() {
	for {
		// Temp
		buf := make([]byte, 0)
		_, sender, err := g.transport.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		go g.handler(buf, sender)
	}
}

func (g *Gossiper) Push(buf []byte) {
	id := idGenerator()
	ch := make(chan Packet)
	g.ackMu.Lock()
	g.ackChan[id] = ch
	g.ackMu.Unlock()

	msg := PushMessage{atomic.AddUint32(&g.seq, 1), id, buf}

	buf, err := AddLabelFromPacket(&msg, g.cfg.encryptType)
	if err != nil {
		// log
		return
	}

	peers := g.SelectRandomPeers(GossipNumber)
	for _, peer := range peers {
		addr, err := net.ResolveUDPAddr("udp", peer)
		if err != nil {
			// log
			continue
		}
		if _, err := g.transport.WriteToUDP(buf, addr); err != nil {
			// log
			continue
		}
	}

	ackCount := 0
	timer := time.NewTimer(300 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case ackPacket := <-ch:
			ack, ok := ackPacket.(*PushAck)
			if !ok {
				// log
				continue
			}
			if ack.Key == id {
				ackCount++
			}
			if ackCount >= (GossipNumber / 2) {
				// correct finish
				return
			}
		case <-timer.C:
			pullSync := PullSync{atomic.LoadUint32(&g.seq), id}

			buf, err := AddLabelFromPacket(&pullSync, g.cfg.encryptType)
			if err != nil {
				// log
				return
			}

			peers := g.SelectRandomPeers(GossipNumber * 2)
			for _, peer := range peers {
				addr, err := net.ResolveUDPAddr("udp", peer)
				if err != nil {
					// log
					continue
				}
				if _, err := g.transport.WriteToUDP(buf, addr); err != nil {
					// log
					continue
				}
			}
		}
	}
}

func (g *Gossiper) MessagePipe() chan []byte {
	return g.messagePipe
}

func (g *Gossiper) SelectRandomPeers(n int) []string {
	return nil
}

func (g *Gossiper) get(key [8]byte) []byte {
	v, ok := g.messageCache.Get(key)
	if !ok {
		return nil
	}
	return v.([]byte)
}
