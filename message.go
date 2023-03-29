package gogossip

import (
	"net"
)

type packet interface {
	Kind() byte

	// 'packet' handler returns packet list if need respond. and basically
	// it is used to divide and transmit a large response. Add 'to' to
	// the packet itself, as there may be times when need to send a
	// response to multiple peers later.
	//
	// It is removed in marshalling and cannot be used on the receiver side.
	To() *net.UDPAddr
}

const (
	pullRequestType  = 0x01
	pullResponseType = 0x02
)

type packetType byte

func (p packetType) String() string {
	switch p {
	case pullRequestType:
		return "PullRequest"
	case pullResponseType:
		return "PullResponse"
	}
	return ""
}

type (
	pullRequest struct {
	}
	pullResponse struct {
		to     *net.UDPAddr
		Keys   [][8]byte
		Values [][]byte
	}
)

func marshalWithEncryption(p packet, encType EncryptType, passphrase string) ([]byte, error) {
	cipher, err := encryptPacket(encType, passphrase, p)
	if err != nil {
		return nil, err
	}
	return bytesToLabel([]byte{p.Kind(), byte(encType)}).combine(cipher)
}

func unmarshalWithDecryption(buf []byte, passphrase string) (*label, []byte, error) {
	label, payload, err := splitLabel(buf)
	if err != nil {
		return nil, nil, err
	}
	plain, err := decryptPayload(payload, EncryptType(label.encryptType), passphrase)
	if err != nil {
		return nil, nil, err
	}
	return label, plain, err
}

func (req *pullRequest) Kind() byte       { return pullRequestType }
func (req *pullRequest) To() *net.UDPAddr { panic("not supported") }

func (res *pullResponse) Kind() byte { return pullResponseType }
func (res *pullResponse) To() *net.UDPAddr {
	if res.to == nil {
		panic("'to' is empty (hint: maybe you are the recipient)")
	}
	return res.to
}
