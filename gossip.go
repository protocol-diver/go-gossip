package gogossip

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"time"
)

const (
	//
	pullInterval = 200 * time.Millisecond

	// Packet fragmentation is not considered.
	maxPacketSize = 65535

	// actualPayloadSize is the result of calculating the overhead
	// in the process of marshaling the PullResponse.
	//
	// TODO(dbadoy): Figure out another serialization method(e.g.
	// gob). The overhead of serializing two dimensional array via
	// json is too much.
	//
	// 	type t struct {
	//		Data [][]byte
	// 	}
	//
	// 	Appending []byte (len: 5000) 1000 times.
	// 	1. json - 6671010
	// 	2. gob  - 5003061
	actualPayloadSize = maxPacketSize - 61440
)

var (
	// If call 'Push' and it returns this error, you are making
	// too many requests. cache is emptied after a certain amount
	// of time, so if you try again, it will be processed normally.
	ErrNoSpaceCache = errors.New("too many requests")
)

type Gossiper struct {
	run    uint32
	cfg    *Config
	logger *log.Logger

	registry  Registry
	transport Transport

	messages *propagator
	pipe     chan []byte
}

func New(reg Registry, transport Transport, cfg *Config) (*Gossiper, error) {
	logger := log.New(os.Stdout, "[Gossip] ", log.LstdFlags)

	if cfg == nil {
		cfg = DefaultConfig()
		logger.Println("config is nil. use default config")
	}
	if cfg.GossipNumber < 2 {
		logger.Println("mininum size of GossipNumber is 2. set default value [2]")
		cfg.GossipNumber = 2
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	filter, err := newFilter(cfg.FilterWithStorage)
	if err != nil {
		return nil, err
	}
	logger.Printf("configured, Filter: %s, GossipNumber: %d, EncryptType: %s\n", filter.Kind(), cfg.GossipNumber, cfg.EncType.String())

	propagator, err := newPropagator(filter)
	if err != nil {
		return nil, err
	}

	gossiper := &Gossiper{
		cfg:       cfg,
		logger:    logger,
		registry:  reg,
		transport: transport,
		messages:  propagator,
		pipe:      make(chan []byte, 4096),
	}
	return gossiper, nil
}

func (g *Gossiper) Start() {
	if atomic.LoadUint32(&g.run) == 1 {
		return
	}
	go g.readLoop()
	go g.pullLoop()
	atomic.StoreUint32(&g.run, 1)
}

// Push is a method for surface to application for starts
// gossip. It's limits requests to prevent abnormal propagation
// when more requests than cacheSize are received.
func (g *Gossiper) Push(buf []byte) error {
	if len(buf) > actualPayloadSize {
		return errors.New("too big")
	}
	if g.messages.size() > cacheSize {
		return ErrNoSpaceCache
	}
	g.push(idGenerator(), buf, false)
	return nil
}

// MessagePipe is a method for surface to application for send
// newly messages.
func (g *Gossiper) MessagePipe() chan []byte {
	return g.pipe
}

func (g *Gossiper) Pending() int {
	return g.messages.size()
}

func (g *Gossiper) readLoop() {
	for {
		buf := make([]byte, maxPacketSize)
		n, sender, err := g.transport.ReadFromUDP(buf)
		if err != nil {
			g.logger.Printf("readLoop: read UDP packet failure, %v\n", err)
			continue
		}

		r := buf[:n]
		g.handler(r, sender)
	}
}

func (g *Gossiper) pullLoop() {
	ticker := time.NewTicker(pullInterval)
	defer ticker.Stop()
	for {
		<-ticker.C

		// Request PullRequest to random peers.
		//
		// Since it is the starting point of the gossip protocol.
		// So it follows the encType of this peer.
		p, err := marshalWithEncryption(&pullRequest{}, g.cfg.EncType, g.cfg.Passphrase)
		if err != nil {
			g.logger.Printf("pullLoop: marshalPacketWithEncryption failure, %v\n", err)
			continue
		}

		// Choose random peers and send.
		multicastWithRawAddress(g.transport, p, g.selectRandomPeers(g.cfg.GossipNumber))
	}
}

func (g *Gossiper) push(key [8]byte, value []byte, remote bool) {
	if g.messages.add(key, value) && remote {
		g.pipe <- value
	}
}

// selectRandomPeers selects random peers.
func (g *Gossiper) selectRandomPeers(n int) []string {
	peers := g.registry.Gossipiers()
	if len(peers) <= n {
		return peers
	}

	var (
		random   = rand.New(rand.NewSource(time.Now().UnixNano()))
		result   = make([]string, 0, n)
		selected = make(map[string]struct{})
	)

	for r := random.Intn(n); len(result) != n; r = random.Intn(n) {
		key := peers[r]
		if _, ok := selected[key]; ok {
			continue
		}
		selected[key] = struct{}{}
		result = append(result, key)
	}

	return result
}

func (g *Gossiper) send(p packet, encType EncryptType) (int, error) {
	b, err := marshalWithEncryption(p, encType, g.cfg.Passphrase)
	if err != nil {
		return 0, err
	}
	return g.transport.WriteToUDP(b, p.To())
}
