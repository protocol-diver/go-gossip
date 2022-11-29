package gogossip

type filter interface {
	Put(key []byte, value []byte) error
	Has(key []byte) bool
	Mod() byte
}
