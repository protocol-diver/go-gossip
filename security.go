package gogossip

import (
	"encoding/json"

	"github.com/protocol-diver/go-gossip/crypto"
)

const (
	NON_SECURE_TYPE = 0x00
	AES256_CBC_TYPE = 0x01
)

type EncryptType byte

func (e EncryptType) String() string {
	switch e {
	case NON_SECURE_TYPE:
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
	case NON_SECURE_TYPE:
		return Cipher{crypto.NON_SECURE{}, kind}
	case AES256_CBC_TYPE:
		return Cipher{crypto.AES256_CBC{}, kind}
	}
	panic("not supported encryption type")
}

func encryptPacket(encType EncryptType, passphrase string, p packet) ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return newCipher(encType).Encrypt(passphrase, b)
}

func decryptPayload(payload []byte, encType EncryptType, passpharse string) ([]byte, error) {
	return newCipher(encType).Decrypt(passpharse, payload)
}
