package gogossip

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Label: packetType(1 byte) + encryptType(1 byte) + actualDataSize(4 byte)

func AddLabelFromPacket(packet Packet, encType EncryptType) ([]byte, error) {
	bpacket, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}
	packetSize := len(bpacket)

	buf := bytes.NewBuffer(make([]byte, 0, 1+1+4+packetSize))

	sizeFlag := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeFlag, uint32(packetSize))

	buf.WriteByte(packet.Kind())
	buf.WriteByte(byte(encType))
	buf.Write(sizeFlag)
	buf.Write(bpacket)

	return buf.Bytes(), nil
}

func RemoveLabelFromPacket(d []byte) ([]byte, byte, EncryptType, error) {
	if len(d) <= 1+1+4 {
		return nil, 0, 0, errors.New("too short")
	}

	buf := bytes.NewBuffer(d)

	packetType, _ := buf.ReadByte()
	encType, _ := buf.ReadByte()

	sizeBuf := make([]byte, 4)
	buf.Read(sizeBuf)
	size := binary.BigEndian.Uint32(sizeBuf)

	data := make([]byte, size)
	n, err := buf.Read(data)
	if err != nil {
		return nil, 0, 0, err
	}
	if n != int(size) {
		return nil, 0, 0, fmt.Errorf("invalid size header: %d, payload: %d", size, n)
	}

	return data, packetType, EncryptType(encType), nil
}

func idGenerator() [8]byte {
	var buf [8]byte
	random := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	random.Read(buf[:])
	return buf
}
