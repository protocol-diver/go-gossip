package gogossip

import (
	"encoding/json"

	"github.com/protocol-diver/go-gossip/crypto"
)

const (
	NO_SECURE_TYPE  = 0x00
	AES256_CBC_TYPE = 0x01
)

type EncryptType byte

func (e EncryptType) String() string {
	switch e {
	case NO_SECURE_TYPE:
		return "NO-SECURE"
	case AES256_CBC_TYPE:
		return "AES256-CBC"
	}
	return ""
}

type CipherMethod interface {
	Encrypt(string, []byte) ([]byte, error)
	Decrypt(string, []byte) ([]byte, error)
}

type Cipher struct {
	CipherMethod
	kind EncryptType
}

func (s *Cipher) Is(kind EncryptType) bool {
	return s.kind == kind
}

func newCipher(kind EncryptType) Cipher {
	switch kind {
	case NO_SECURE_TYPE:
		return Cipher{NO_SECURE{}, kind}
	case AES256_CBC_TYPE:
		return Cipher{crypto.AES256_CBC{}, kind}
	}
	panic("not supported encryption type")
}

func EncryptPacket(encType EncryptType, passphrase string, packet Packet) ([]byte, error) {
	b, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}
	cipher, err := newCipher(encType).Encrypt(passphrase, b)
	if err != nil {
		return nil, err
	}
	return cipher, err
}

func DecryptPayload(encType EncryptType, passpharse string, payload []byte) ([]byte, error) {
	plain, err := newCipher(encType).Decrypt(passpharse, payload)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

//
type NO_SECURE struct{}

func (n NO_SECURE) Encrypt(passphrase string, buf []byte) ([]byte, error) {
	return buf, nil
}

func (n NO_SECURE) Decrypt(passphrase string, buf []byte) ([]byte, error) {
	return buf, nil
}
