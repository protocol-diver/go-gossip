package gogossip

import (
	"crypto/rsa"

	"github.com/dbadoy/go-gossip/crypto"
)

const (
	TEMP_NONE_ENC   = 0x00
	AES256_CBC_TYPE = 0x01
)

type EncryptType byte

func (e EncryptType) String() string {
	switch e {
	case TEMP_NONE_ENC:
		return "N"
	case AES256_CBC_TYPE:
		return "AES256-CBC"
	}
	return ""
}

type CipherMethod interface {
	// SymmetricCipher
	Encrypt(string, []byte) ([]byte, error)
	Decrypt(string, []byte) ([]byte, error)

	// ?
	EncryptWithPublicKey(*rsa.PublicKey, []byte) ([]byte, error)
	DecryptWithPrivateKey(*rsa.PrivateKey, []byte) ([]byte, error)
}

type Cipher struct {
	CipherMethod
	kind EncryptType
}

func (s *Cipher) Is(kind EncryptType) bool {
	return s.kind == kind
}

func NewCipher(kind EncryptType) Cipher {
	switch kind {
	case AES256_CBC_TYPE:
		return Cipher{crypto.AES256_CBC{}, kind}
	}
	panic("not supported encryption type")
}
