package gogossip

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
)

// filter is a data structure to check whether a message
// has been received. The library should ignore messages
// that have already been received and only accept new
// messages. In cases sensitive to duplicate messages,
// storage is used to maintain filter even when the node
// is restarted. Memory filter can be used if you are not
// very sensitive to duplicate messages.
type filter interface {
	// Registers received messages in filter. Only need to enter
	// the key of the received message in the Put parameter. The
	// corresponding Value is stored as nil (only check if it has
	// been received).
	Put(key []byte) error

	Has(key []byte) bool
	Mod() string
}

func newFilter(filterWithStorage string) (filter, error) {
	if filterWithStorage == "" {
		// bloom filter
	} else {
		return newStorageFilter(filterWithStorage)
	}
	return nil, errors.New("non")
}

// level db
type storageFilter struct {
	db *leveldb.DB
}

func newStorageFilter(path string) (*storageFilter, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &storageFilter{db}, nil
}

func (s *storageFilter) Put(key []byte) error {
	return s.db.Put(key, nil, nil)
}

func (s *storageFilter) Has(key []byte) bool {
	has, err := s.db.Has(key, nil)
	if err != nil {
		panic(err)
	}
	return has
}

func (*storageFilter) Mod() string { return "LevelDB" }

// bloom filter
type memoryFilter struct {
	//
}

func (*memoryFilter) Mod() string { return "BloomFilter" }
