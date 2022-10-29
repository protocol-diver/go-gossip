package gogossip

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	// byte
	packetTypeFlagSize = 1
	encryptFlagSize    = 1 // binary.Size((EncryptType(0)))

	// uint32, This is not necessary unless you add a specific flag (eg checksum) after the data.
	actualDataFlagSize = 4
)

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
	label[packetTypeFlagSize] = packet.Kind()
	label[encryptFlagSize] = byte(encType)

	// Use copy beacuase this process is size-sensitive.
	copy(label[encryptFlagSize+1:actualDataFlagSize], sizeFlag)

	r := make([]byte, len(label)+packetSize)
	copy(r, label)
	r = append(r, bpacket...)
	return r, nil
}

func RemoveLabelFromPacket(d []byte) ([]byte, byte, EncryptType, error) {
	if len(d) <= encryptFlagSize+actualDataFlagSize {
		return nil, 0, 0, errors.New("too short")
	}
	sizeBuf := d[encryptFlagSize+1 : actualDataFlagSize]
	size := binary.BigEndian.Uint32(sizeBuf)

	// Logic to get postfix flag value . . .
	//
	//
	data := d[actualDataFlagSize+1:]

	packetType := d[packetTypeFlagSize]
	encType := d[encryptFlagSize]

	if len(data) != int(size) {
		return nil, 0, 0, fmt.Errorf("invalid size, label: %d, packet: %d", size, len(data))
	}
	return data, packetType, EncryptType(encType), nil
}
