package gogossip

import (
	"testing"
	"time"
)

func TestDuplicateAdd(t *testing.T) {
	f, err := newFilter("")
	if err != nil {
		t.Fatal(err)
	}
	prop, err := newPropagator(f)
	if err != nil {
		t.Fatal(err)
	}

	prop.add([8]byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{1})

	if prop.add([8]byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{1}) {
		t.Fatal("TestDuplicateAdd failure, occured duplicate add")
	}
	if !prop.add([8]byte{0, 0, 0, 0, 0, 0, 0, 0}, []byte{1}) {
		t.Fatal("TestDuplicateAdd failure, failed newly message add")
	}
}

func TestClearItems(t *testing.T) {
	f, err := newFilter("")
	if err != nil {
		t.Fatal(err)
	}
	prop, err := newPropagator(f)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < cacheSize; i++ {
		prop.add(idGenerator(), []byte{1})
	}

	// touch
	prop.items()

	time.Sleep(time.Second)
	if after := prop.size(); after != 0 {
		t.Fatalf("TestClearItems failure, want: %v, got: %v", 0, after)
	}
}
