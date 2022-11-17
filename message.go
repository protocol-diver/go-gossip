package gogossip

import (
	"net"
)

type Packet interface {
	Kind() byte

	// Packet handler returns packet list if need respond. and basically
	// it is used to divide and transmit a large response. Add 'to' to
	// the packet itself, as there may be times when need to send a
	// response to multiple peers later.
	//
	// It is removed in marshalling and cannot be used on the receiver side.
	To() *net.UDPAddr
}

const (
	PullRequestType  = 0x01
	PullResponseType = 0x02
)

type (
	PullRequest struct {
	}
	PullResponse struct {
		to     *net.UDPAddr
		Keys   [][8]byte
		Values [][]byte
	}
)

func marshalPacketWithEncryption(packet Packet, encType EncryptType, passphrase string) ([]byte, error) {
	cipher, err := EncryptPacket(encType, passphrase, packet)
	if err != nil {
		return nil, err
	}
	return bytesToLabel([]byte{packet.Kind(), byte(encType)}).combine(cipher)
}

func (req *PullRequest) Kind() byte       { return PullRequestType }
func (req *PullRequest) To() *net.UDPAddr { panic("not supported") }

func (res *PullResponse) Kind() byte { return PullResponseType }
func (res *PullResponse) To() *net.UDPAddr {
	if res.to == nil {
		panic("'to' is empty (hint: maybe you are the recipient)")
	}
	return res.to
}
