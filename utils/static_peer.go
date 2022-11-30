package utils

import "sync"

type RegistryGossip struct {
	mu sync.Mutex
	m  map[string]string
}

func NewRegistryGossip() RegistryGossip {
	return RegistryGossip{
		m: make(map[string]string),
	}
}

func (r *RegistryGossip) Register(k, v string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[k] = v
}

func (r *RegistryGossip) Deregister(k string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.m, k)
}

func (r *RegistryGossip) Gossipiers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := make([]string, 0, len(r.m))
	for _, raw := range r.m {
		res = append(res, raw)
	}
	return res
}

func (r *RegistryGossip) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.m)
}
