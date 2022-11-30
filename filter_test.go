package gogossip

import (
	"os"
	"testing"
)

func TestRouteFilter(t *testing.T) {
	tds := []struct {
		filterWithStorage string
		want              string
	}{
		{
			filterWithStorage: "",
			want:              "BloomFilter",
		},
		{
			filterWithStorage: "./temp",
			want:              "LevelDB",
		},
	}

	for _, td := range tds {
		filter, err := newFilter(td.filterWithStorage)
		if err != nil {
			t.Fatal(err)
		}
		if td.want != filter.Mod() {
			t.Fatalf("RouteFilter test failure, want: %s, got: %s", td.want, filter.Mod())
		}

		// remove the directory if filter mod is storage.
		if filter.Mod() == "LevelDB" {
			os.RemoveAll(td.filterWithStorage)
		}
	}
}

func TestMemoryFilter(t *testing.T) {
	tds := []struct {
		data  []byte
		query []byte
		want  bool
	}{
		{
			data:  []byte{1, 2, 3},
			query: []byte{1, 2, 3},
			want:  true,
		},
		{
			data:  []byte{4, 5, 6},
			query: []byte{1},
			want:  false,
		},
	}

	filter, err := newFilter("")
	if err != nil {
		t.Fatal(err)
	}

	for _, td := range tds {
		if err := filter.Put(td.data); err != nil {
			t.Fatal(err)
		}

		if filter.Has(td.query) != td.want {
			t.Fatalf("MemoryFilter test failure, put: %v, query: %v, want: %v, got: %v", td.data, td.query, td.want, filter.Has(td.query))
		}
	}
}

func TestStorageFilter(t *testing.T) {
	tds := []struct {
		data  []byte
		query []byte
		want  bool
	}{
		{
			data:  []byte{1, 2, 3},
			query: []byte{1, 2, 3},
			want:  true,
		},
		{
			data:  []byte{4, 5, 6},
			query: []byte{1},
			want:  false,
		},
	}

	path := "./temp"

	filter, err := newFilter(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, td := range tds {
		if err := filter.Put(td.data); err != nil {
			t.Fatal(err)
		}

		if filter.Has(td.query) != td.want {
			t.Fatalf("StorageFilter test failure, put: %v, query: %v, want: %v, got: %v", td.data, td.query, td.want, filter.Has(td.query))
		}
	}

	os.RemoveAll(path)
}
