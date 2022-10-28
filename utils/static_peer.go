package utils

type StaticPeers map[string]string

type RegistryGossip map[string]string

func (r *RegistryGossip) Register(k, v string) {
	(*r)[k] = v
}

func (r *RegistryGossip) Deregister(k string) {
	delete((*r), k)
}

func (r *RegistryGossip) Gossipiers() []string {
	res := make([]string, len(*r))
	for _, raw := range *r {
		res = append(res, raw)
	}
	return res
}

func (r *RegistryGossip) Size() int {
	return len(*r)
}
