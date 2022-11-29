package gogossip

import (
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type broadcast struct {
	c *lru.Cache
	f filter
}

func newBroadcast(f filter) (*broadcast, error) {
	cache, err := lru.New(256)
	if err != nil {
		return nil, err
	}

	return &broadcast{cache, f}, nil
}

func (b *broadcast) add(key [8]byte, value []byte) bool {
	if has := b.f.Has(key[:]); has {
		return false
	}
	if contain := b.c.Contains(key); contain {
		return false
	}

	b.c.Add(key, value)

	if err := b.f.Put(key[:], nil); err != nil {
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

			go func(k interface{}) {
				time.Sleep(1000 * time.Millisecond)
				b.c.Remove(k)
			}(key)
		}
	}
	return kl, vl
}

func (b *broadcast) size() int {
	return b.c.Len()
}
