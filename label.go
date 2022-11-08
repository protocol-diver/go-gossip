package gogossip

import (
	"encoding/binary"
	"math/rand"
	"time"
)

// Label: packetType(1 byte) + encryptType(1 byte)

var labelTotalSize = binary.Size(Label{})

type Label struct {
	packetType  byte
	encryptType byte
	// tempSize    [4]byte
}

func (l Label) combine(buf []byte) []byte {
	b := make([]byte, 0, labelTotalSize+len(buf))
	b = append(b, l.bytes()...)
	b = append(b, buf...)
	return b
}

func (l Label) bytes() []byte {
	b := make([]byte, labelTotalSize)
	b[0] = l.packetType
	b[1] = l.encryptType
	// copy(b[2:6], l.tempSize[:])
	return b
}

func BytesToLabel(buf []byte) Label {
	if len(buf) != labelTotalSize {
		panic("hi")
	}
	// var tempSize [4]byte
	// copy(tempSize[:], buf[2:4])

	return Label{
		buf[0], buf[1], // tempSize,
	}
}

func SplitLabel(buf []byte) (Label, []byte) {
	if len(buf) < labelTotalSize {
		panic("hi")
	}
	return BytesToLabel(buf[:labelTotalSize]), buf[labelTotalSize:]
}

func idGenerator() [8]byte {
	var buf [8]byte
	random := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	random.Read(buf[:])
	return buf
}
