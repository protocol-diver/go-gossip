package gogossip

import (
	"github.com/VictoriaMetrics/fastcache"
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
	// Put registers received messages in filter. Only need to
	// enter the key of the received message in the Put parameter.
	// The corresponding Value is stored as nil (only check if it
	// has been received).
	Put(key []byte) error

	// Has checks if that key exists in filter and return whether
	// or not.
	Has(key []byte) bool

	// Kind is simply method for logging. It should return which
	// filter implementation it is.
	Kind() string
}

func newFilter(filterWithStorage string) (filter, error) {
	if filterWithStorage == "" {
		return newMemoryFilter()
	} else {
		return newStorageFilter(filterWithStorage)
	}
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

func (*storageFilter) Kind() string { return "LevelDB" }

type memoryFilter struct {
	f *fastcache.Cache
}

func newMemoryFilter() (*memoryFilter, error) {
	// 32 MiB
	return &memoryFilter{fastcache.New(32 << 20)}, nil
}

func (m *memoryFilter) Put(key []byte) error {
	m.f.Set(key, nil)
	return nil
}

func (m *memoryFilter) Has(key []byte) bool {
	return m.f.Has(key)
}

func (*memoryFilter) Kind() string { return "FastCache" }
