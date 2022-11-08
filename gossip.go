package gogossip

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	//
	gossipNumber = 3
	// Set to twice the predicted value of messages to occur per second
	messageCacheSize = 512
	//
	messagePipeSize = 4096
	//
	maxPacketSize = 8 * 1024
	//
	ackTimeout = 300 * time.Millisecond
)

type Gossiper struct {
	run uint32
	seq uint32

	discovery Discovery
	transport Transport

	messagePipe chan []byte

	ackChan map[[8]byte]chan PushAck
	ackMu   sync.Mutex

	messageCache *lru.Cache

	cfg *Config
}

func NewGossiper(discv Discovery, transport Transport, cfg *Config) (*Gossiper, error) {
	cache, err := lru.New(512)
	if err != nil {
		return nil, err
	}
	gossiper := &Gossiper{
		seq:          0,
		discovery:    discv,
		transport:    transport,
		messagePipe:  make(chan []byte, 4096),
		ackChan:      make(map[[8]byte]chan PushAck),
		messageCache: cache,
		cfg:          cfg,
	}
	return gossiper, nil
}

func (g *Gossiper) Start() {
	if atomic.LoadUint32(&g.run) == 1 {
		return
	}
	go g.dispatch()
}

func (g *Gossiper) dispatch() {
	for {
		// Temp
		buf := make([]byte, 8192)
		_, sender, err := g.transport.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		go g.handler(buf, sender)
	}
}

func (g *Gossiper) Push(buf []byte) {
	// Make new gossip message id.
	id := idGenerator()

	// Allocate channel for receive the ACKs from PushMessage.
	ch := make(chan PushAck)

	g.ackMu.Lock()
	g.ackChan[id] = ch
	g.ackMu.Unlock()

	// Deallocate ACK channel when AckTimeout reached.
	defer func() {
		g.ackMu.Lock()
		delete(g.ackChan, id)
		g.ackMu.Unlock()
	}()

	msg := &PushMessage{atomic.AddUint32(&g.seq, 1), id, buf}

	// Encryption.
	cipher, err := EncryptPacket(g.cfg.encryptType, g.cfg.passphrase, msg)
	if err != nil {
		return
	}
	// Combine Label with Payload.
	p := BytesToLabel([]byte{msg.Kind(), byte(g.cfg.encryptType)}).combine(cipher)

	//
	multicastWithRawAddress(g.transport, g.SelectRandomPeers(gossipNumber), p)

	// Starts count ACK messages.
	ackCount := 0
	timer := time.NewTimer(300 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case ack := <-ch:
			if ack.Key == id {
				ackCount++
			}
			// Normal finish
			if ackCount >= (gossipNumber / 2) {
				return
			}
		case <-timer.C:
			// Send PullSync for starts the Pull flow if timeout reached.
			pullSync := &PullSync{atomic.LoadUint32(&g.seq), id}

			cipher, err := EncryptPacket(g.cfg.encryptType, g.cfg.passphrase, pullSync)
			if err != nil {
				return
			}
			p := BytesToLabel([]byte{pullSync.Kind(), byte(g.cfg.encryptType)}).combine(cipher)

			multicastWithRawAddress(g.transport, g.SelectRandomPeers(gossipNumber*2), p)
		}
	}
}

//
func (g *Gossiper) MessagePipe() chan []byte {
	return g.messagePipe
}

//
func (g *Gossiper) SelectRandomPeers(n int) []string {
	peers := g.discovery.Gossipiers()
	if len(peers) <= n {
		return peers
	}
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	// TODO(dbadoy): Avoid duplicate selection.
	selected := make([]string, 0, n)
	for i := 0; i < n; i++ {
		selected = append(selected, peers[random.Intn(n)])
	}
	return selected
}

//
func (g *Gossiper) get(key [8]byte) []byte {
	v, ok := g.messageCache.Get(key)
	if !ok {
		return nil
	}
	return v.([]byte)
}

//
func (g *Gossiper) add(key [8]byte, value []byte) {
	g.messageCache.Add(key, value)
	go func() {
		g.messagePipe <- value
	}()
}
