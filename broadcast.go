package gogossip

import (
	"sync"
	"time"
)

const (
	timeout = 30 * time.Second
)

// messages stores recent gossip messages.
// It also stores a deadline for deletion from memory.
type message struct {
	value    []byte
	deadline time.Time

	// Marking for the requestor. Exclude if the requestor has already taken it.
	// If don't check it here, occur a unnecessary response.
	touched map[string]bool
}

// broadcast manages the message in the form of a map.
// The message exist here means a message to be propagated to
// neighboring nodes.
type broadcast struct {
	mu sync.Mutex
	m  map[[8]byte]message
}

// add is a method for storing messages in broadcast.
// If it already exists, it is skipped, and if it does not exist,
// it is added by setting a deadline.
func (b *broadcast) add(key [8]byte, value []byte) bool {
	b.mu.Lock()
	if _, ok := b.m[key]; ok {
		// This message already received.
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

// keys returns all keys that exist in the map in broadcast.
//
// The caller must hold b.mu.
func (b *broadcast) keys() [][8]byte {
	keys := make([][8]byte, 0, len(b.m))
	for k := range b.m {
		keys = append(keys, k)
	}
	return keys
}

// itemsWithTouch gets all messages that exist in B. In the process,
// messages already taken by the requester are excluded. It also marks
// the returned message(to exclude next request).
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

// timeoutLoop periodically finds and deletes message that has expired in broadcast.
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
