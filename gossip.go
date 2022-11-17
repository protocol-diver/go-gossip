package gogossip

import (
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	//
	gossipNumber = 2
	//
	pullInterval = 200 * time.Millisecond
	//
	actualDataSize = 512
)

type Gossiper struct {
	run uint32

	discovery Discovery
	transport Transport

	messages broadcast
	pipe     chan []byte

	cfg *Config
}

func NewGossiper(discv Discovery, transport Transport, cfg *Config) (*Gossiper, error) {
	if cfg == nil {
		cfg = &Config{
			encryptType: NON_SECURE_TYPE,
			passphrase:  "",
		}
	}
	gossiper := &Gossiper{
		discovery: discv,
		transport: transport,
		messages: broadcast{
			m: make(map[[8]byte]message),
		},
		pipe: make(chan []byte, actualDataSize),
		cfg:  cfg,
	}
	return gossiper, nil
}

func (g *Gossiper) Start() {
	if atomic.LoadUint32(&g.run) == 1 {
		return
	}
	go g.messages.timeoutLoop()
	go g.readLoop()
	go g.pullLoop()
	//
	atomic.StoreUint32(&g.run, 1)
}

func (g *Gossiper) pullLoop() {
	ticker := time.NewTicker(pullInterval)
	defer ticker.Stop()
	for {
		<-ticker.C

		// Request PullRequest to random peers.
		msg := &PullRequest{}

		// Since it is the starting point of the gossip protocol.
		// So it follows the encType of this peer.
		p, err := marshalPacketWithEncryption(msg, g.cfg.encryptType, g.cfg.passphrase)
		if err != nil {
			log.Printf("pullLoop: marshalPacketWithEncryption failure %v", err)
			continue
		}

		// Choose random peers and send.
		multicastWithRawAddress(g.transport, p, g.selectRandomPeers(gossipNumber))
	}
}

func (g *Gossiper) readLoop() {
	for {
		// Temprary packet limit. Need basis.
		buf := make([]byte, 8192)
		n, sender, err := g.transport.ReadFromUDP(buf)
		if err != nil {
			log.Printf("readLoop: read UDP packet failure %v", err)
			continue
		}

		// Slice actual data.
		r := buf[:n]
		go g.handler(r, sender)
	}
}

// Surface to application for starts gossip.
func (g *Gossiper) Push(buf []byte) {
	g.push(idGenerator(), buf)
}

// Surface to application for send newly messages.
func (g *Gossiper) MessagePipe() chan []byte {
	return g.pipe
}

func (g *Gossiper) push(key [8]byte, value []byte) {
	if g.messages.add(key, value) {
		g.pipe <- value
	}
}

// Select random peers.
func (g *Gossiper) selectRandomPeers(n int) []string {
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

func (g *Gossiper) send(packet Packet, encType EncryptType) error {
	b, err := marshalPacketWithEncryption(packet, encType, g.cfg.passphrase)
	if err != nil {
		return err
	}
	_, err = g.transport.WriteToUDP(b, packet.To())
	return err
}
