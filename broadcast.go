package gogossip

import (
	"sync"
	"time"
)

const (
	//
	timeout = 30 * time.Second
)

type message struct {
	value    []byte
	deadline time.Time

	// Marking for the requestor. Exclude if the requestor has already taken it.
	touched map[string]bool
}

type broadcast struct {
	mu sync.Mutex
	m  map[[8]byte]message
}

func (b *broadcast) add(key [8]byte, value []byte) bool {
	b.mu.Lock()
	if _, ok := b.m[key]; ok {
		// already received
		b.mu.Unlock()
		return false
	}
	b.m[key] = message{
		value:    value,
		deadline: time.Now().Add(timeout),
		touched:  make(map[string]bool),
	}
	b.mu.Unlock()

	return true
}

// The caller must hold b.mu.
func (b *broadcast) keys() [][8]byte {
	keys := make([][8]byte, 0, len(b.m))
	for k := range b.m {
		keys = append(keys, k)
	}
	return keys
}

func (b *broadcast) itemsWithTouch(addr string) ([][8]byte, [][]byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	keys := b.keys()

	rk := make([][8]byte, 0, len(keys))
	rv := make([][]byte, 0, len(keys))

	for k, v := range b.m {
		if !v.touched[addr] {
			rk = append(rk, k)
			rv = append(rv, v.value)
		}
		v.touched[addr] = true
	}
	return rk, rv
}

func (b *broadcast) timeoutLoop() {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	for {
		<-ticker.C

		now := time.Now()
		b.mu.Lock()
		keys := b.keys()
		for _, key := range keys {
			if _, ok := b.m[key]; !ok {
				return
			}
			if b.m[key].deadline.Before(now) {
				delete(b.m, key)
			}
		}
		b.mu.Unlock()
	}
}
