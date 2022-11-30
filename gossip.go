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
	pullInterval = 100 * time.Millisecond

	// 2^18(65535) - ip header(20) - udp header(8)
	maxPacketSize = 65507

	// actualPayloadSize is the result of calculating the overhead
	// in the process of marshaling the PullResponse.
	actualPayloadSize = maxPacketSize - 61440
)

type Gossiper struct {
	run    uint32
	cfg    *Config
	logger *log.Logger

	discovery Discovery
	transport Transport

	messages *broadcast
	pipe     chan []byte
}

func New(discv Discovery, transport Transport, cfg *Config) (*Gossiper, error) {
	logger := log.New(os.Stdout, "[Gossip] ", log.LstdFlags)

	if cfg == nil {
		cfg = DefaultConfig()
		logger.Println("config is nil. use default config")
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if cfg.GossipNumber < 2 {
		logger.Println("mininum size of GossipNumber is 2. set default value [2]")
		cfg.GossipNumber = 2
	}
	filter, err := newFilter(cfg.FilterWithStorage)
	if err != nil {
		return nil, err
	}
	logger.Printf("configured, FilterMod: %s, GossipNumber: %d, EncryptType: %s\n", filter.Mod(), cfg.GossipNumber, cfg.EncType.String())

	broadcast, err := newBroadcast(filter)
	if err != nil {
		return nil, err
	}

	gossiper := &Gossiper{
		cfg:       cfg,
		logger:    logger,
		discovery: discv,
		transport: transport,
		messages:  broadcast,
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
		return errors.New("too many requests")
	}
	g.push(idGenerator(), buf, false)
	return nil
}

// MessagePipe is a method for surface to application for send
// newly messages.
func (g *Gossiper) MessagePipe() chan []byte {
	return g.pipe
}

func (g *Gossiper) Size() int {
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
		go g.handler(r, sender)
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
		p, err := marshalWithEncryption(&PullRequest{}, g.cfg.EncType, g.cfg.Passphrase)
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
	peers := g.discovery.Gossipiers()
	if len(peers) <= n {
		return peers
	}

	var (
		random   = rand.New(rand.NewSource(time.Now().UnixNano()))
		indices  = make([]int, 0, n)
		selected = make([]string, 0, n)
	)

	// Avoid duplicate selection.
	for r := random.Intn(n); len(indices) != n; r = random.Intn(n) {
		for _, v := range indices {
			if v == r {
				continue
			}
		}
		indices = append(indices, r)
	}

	for i := 0; i < n; i++ {
		selected = append(selected, peers[indices[i]])
	}

	return selected
}

func (g *Gossiper) send(packet Packet, encType EncryptType) (int, error) {
	b, err := marshalWithEncryption(packet, encType, g.cfg.Passphrase)
	if err != nil {
		return 0, err
	}
	return g.transport.WriteToUDP(b, packet.To())
}
