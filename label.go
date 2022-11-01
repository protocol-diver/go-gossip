package gogossip

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var (
	// byte
	packetTypeFlagSize = 1
	encryptFlagSize    = 1 // binary.Size((EncryptType(0)))
	actualDataFlagSize = 4 // uint32, This is not necessary unless you add a specific flag (eg checksum) after the data.

	// TODO: not good
	packetTypeFlag = 0
	encryptFlag    = packetTypeFlag + packetTypeFlagSize
	actualDataFlag = encryptFlag + encryptFlagSize
)

// Use bytes.Buffer ?
func AddLabelFromPacket(packet Packet, encType EncryptType) ([]byte, error) {
	bpacket, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}
	packetSize := len(bpacket)

	sizeFlag := make([]byte, actualDataFlagSize)
	binary.BigEndian.PutUint32(sizeFlag, uint32(packetSize))

	label := make([]byte, packetTypeFlagSize+encryptFlagSize+actualDataFlagSize)

	// packetTypeFlag, encryptFlag is always 1 byte.
	label[packetTypeFlag] = packet.Kind()
	label[encryptFlag] = byte(encType)

	// Use copy beacuase this process is size-sensitive.
	copy(label[actualDataFlag:actualDataFlagSize], sizeFlag)

	r := make([]byte, len(label)+packetSize)
	copy(r, label)
	r = append(r, bpacket...)
	return r, nil
}

func RemoveLabelFromPacket(d []byte) ([]byte, byte, EncryptType, error) {
	if len(d) <= encryptFlagSize+actualDataFlagSize {
		return nil, 0, 0, errors.New("too short")
	}
	sizeBuf := d[encryptFlag+1 : actualDataFlag]
	size := binary.BigEndian.Uint32(sizeBuf)

	// Logic to get postfix flag value . . .
	//
	//
	data := d[actualDataFlag+1:]

	packetType := d[packetTypeFlag]
	encType := d[encryptFlag]

	if len(data) != int(size) {
		return nil, 0, 0, fmt.Errorf("invalid size, label: %d, packet: %d", size, len(data))
	}
	return data, packetType, EncryptType(encType), nil
}

func idGenerator() [8]byte {
	var buf [8]byte
	random := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	random.Read(buf[:])
	return buf
}
