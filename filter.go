package gogossip

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
)

type filter interface {
	Put(key []byte, value []byte) error
	Has(key []byte) bool
	Mod() string
}

func newFilter(filterWithStorage string) (filter, error) {
	if filterWithStorage == "" {
		// bloom filter
	} else {
		db, err := leveldb.OpenFile(filterWithStorage, nil)
		if err != nil {
			return nil, err
		}
		return &storage{db}, nil
	}
	return nil, errors.New("non")
}

// level db
type storage struct {
	db *leveldb.DB
}

func (s *storage) Put(key []byte, value []byte) error {
	return s.db.Put(key, value, nil)
}

func (s *storage) Has(key []byte) bool {
	has, err := s.db.Has(key, nil)
	if err != nil {
		panic(err)
	}
	return has
}

func (*storage) Mod() string { return "LevelDB" }

// bloom filter
type memory struct {
	//
}
