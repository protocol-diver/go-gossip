package gogossip

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"time"
)

var (
	errTooShort = errors.New("bytes too short")
)

// Label: packetType(1 byte) + encryptType(1 byte)

var labelTotalSize = binary.Size(label{})

type label struct {
	packetType  byte
	encryptType byte
	// tempSize    [4]byte
}

func (l label) combine(buf []byte) ([]byte, error) {
	if binary.Size(l) != labelTotalSize {
		return nil, errTooShort
	}
	b := make([]byte, 0, labelTotalSize+len(buf))
	b = append(b, l.bytes()...)
	b = append(b, buf...)
	return b, nil
}

func (l label) bytes() []byte {
	b := make([]byte, labelTotalSize)
	b[0] = l.packetType
	b[1] = l.encryptType
	// copy(b[2:6], l.tempSize[:])
	return b
}

func bytesToLabel(buf []byte) *label {
	if len(buf) != labelTotalSize {
		return nil
	}
	// var tempSize [4]byte
	// copy(tempSize[:], buf[2:4])

	return &label{
		buf[0], buf[1], // tempSize,
	}
}

func splitLabel(buf []byte) (*label, []byte, error) {
	if len(buf) < labelTotalSize {
		return nil, nil, errTooShort
	}
	return bytesToLabel(buf[:labelTotalSize]), buf[labelTotalSize:], nil
}

func idGenerator() [8]byte {
	var buf [8]byte
	random := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	random.Read(buf[:])
	return buf
}
