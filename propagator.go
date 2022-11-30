package gogossip

import (
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	cacheSize = 512
)

// propagator is a data structure required for peers to use to
// propagate. propagator manages new messages to propagate to
// other peers; determine whether the message is received or not.
// If it is a new message, it is stored in cache and registered
// in filter. Ignore messages that have already been received.
type propagator struct {
	c *lru.Cache // TODO: temp impl. Need impl MFU
	f filter
}

func newPropagator(f filter) (*propagator, error) {
	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}
	return &propagator{
		c: cache,
		f: f,
	}, nil
}

func (p *propagator) add(key [8]byte, value []byte) bool {
	// It is skipped if the corresponding key exists
	// in the filter or cache.
	//
	// hold mu?
	if has := p.f.Has(key[:]); has {
		return false
	}
	if contain := p.c.Contains(key); contain {
		return false
	}

	// Register in the filter and saving the value in the cache.
	p.c.Add(key, value)
	if err := p.f.Put(key[:]); err != nil {
		panic(err)
	}

	return true
}

func (p *propagator) items() ([][8]byte, [][]byte) {
	kl := make([][8]byte, 0)
	vl := make([][]byte, 0)

	keys := p.c.Keys()
	for _, key := range keys {
		if value, ok := p.c.Get(key); ok {
			kl = append(kl, key.([8]byte))
			vl = append(vl, value.([]byte))

			// The data in the cache is removed after performing
			// pullInterval 5 times. (Based on Best Effort that
			// it would have spread evenly after 5 times of
			// propagation)
			go func(k interface{}) {
				time.Sleep(5 * pullInterval)
				p.c.Remove(k)
			}(key)
		}
	}
	return kl, vl
}

func (p *propagator) size() int {
	return p.c.Len()
}
