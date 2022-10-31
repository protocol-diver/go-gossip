package gogossip

// OR use name 'Registration'
type Discovery interface {
	// The imported DHT object must have method Gossipiers.
	// It returns all of the peer that target of gossip protocol by array of
	// raw addresses.
	// The raw address validate when before send gossip.
	//
	// It'll used 'SelectRandomPeers' in Gossiper.
	Gossipiers() []string
}
