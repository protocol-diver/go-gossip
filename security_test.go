package gogossip

import (
	"errors"
	"testing"
)

func TestNewCipher(t *testing.T) {
	tds := []struct {
		eType EncryptType
		err   error
	}{
		{
			eType: NON_SECURE_TYPE,
			err:   nil,
		},
		{
			eType: AES256_CBC_TYPE,
			err:   nil,
		},
		{
			eType: 0x03,
			err:   errors.New("not supported encryption type"),
		},
	}

	for _, td := range tds {
		defer func() {
			if err := recover(); err != nil {
				if td.err == nil {
					t.Fatalf("TestNewCipher failure, want: %v, got: %v", td.err, err)
				} else {
					if err != td.err.Error() {
						t.Fatalf("TestNewCipher failure, want: %v, got: %v", td.err, err)
					}
				}
			}
		}()
		newCipher(EncryptType(td.eType))
	}
}

func TestEncrypt_AES256_CBC(t *testing.T) {
	tds := []struct {
		ekey  string
		dkey  string
		plain string
		err   bool
	}{
		{ekey: "seungbae", dkey: "seungbae", plain: "hello, world", err: false},
		{ekey: "seungbae", dkey: "seungbae1", plain: "hello, world", err: true},
	}

	cipher := newCipher(AES256_CBC_TYPE)

	for _, td := range tds {
		enc, err := cipher.Encrypt(td.ekey, []byte(td.plain))
		if err != nil {
			t.Fatal(err)
		}
		dec, err := cipher.Decrypt(td.dkey, enc)
		if err != nil {
			if !td.err {
				t.Fatal(err)
			}
		}

		if !td.err {
			if string(dec) != td.plain {
				t.Fatalf("TestEncrypt_AES256_CBC failure, want: %s, got: %s", string(td.plain), string(dec))
			}
		}
	}
}
