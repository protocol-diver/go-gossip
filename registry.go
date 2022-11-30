package gogossip

type Registry interface {
	// The imported Registry object must have method Gossipiers implement thread
	// safety. It returns all of the peer that target of gossip protocol by
	// array of raw addresses. The raw address validate when before send gossip.
	//
	// It'll used 'SelectRandomPeers' in Gossiper.
	Gossipiers() []string
}
