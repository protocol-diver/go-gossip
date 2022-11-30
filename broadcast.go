package gogossip

import (
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	cacheSize = 512
)

type broadcast struct {
	c *lru.Cache // TODO: temp impl. Need impl MFU
	f filter
}

func newBroadcast(f filter) (*broadcast, error) {
	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}
	return &broadcast{
		c: cache,
		f: f,
	}, nil
}

func (b *broadcast) add(key [8]byte, value []byte) bool {
	// It is skipped if the corresponding key exists
	// in the filter or cache.
	//
	// hold mu?
	if has := b.f.Has(key[:]); has {
		return false
	}
	if contain := b.c.Contains(key); contain {
		return false
	}

	// Register in the filter and saving the value in the cache.
	b.c.Add(key, value)
	if err := b.f.Put(key[:]); err != nil {
		panic(err)
	}

	return true
}

func (b *broadcast) items() ([][8]byte, [][]byte) {
	kl := make([][8]byte, 0)
	vl := make([][]byte, 0)

	keys := b.c.Keys()
	for _, key := range keys {
		if value, ok := b.c.Get(key); ok {
			kl = append(kl, key.([8]byte))
			vl = append(vl, value.([]byte))

			// The data in the cache is removed after performing
			// pullInterval 5 times. (Based on Best Effort that
			// it would have spread evenly after 5 times of
			// propagation)
			go func(k interface{}) {
				time.Sleep(5 * pullInterval)
				b.c.Remove(k)
			}(key)
		}
	}
	return kl, vl
}

func (b *broadcast) size() int {
	return b.c.Len()
}
