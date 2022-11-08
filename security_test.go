package gogossip

import (
	"fmt"
	"testing"
)

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
		defer func() {
			err := recover()
			if !td.err && err != nil {
				fmt.Println(td.err, err)
				t.Fatal("fil")
			}
		}()
		enc, err := cipher.Encrypt(td.ekey, []byte(td.plain))
		if err != nil {
			t.Fatal(err)
		}
		dec, err := cipher.Decrypt(td.dkey, enc)
		if err != nil {
			t.Fatal(err)
		}
		if string(dec) != td.plain {
			panic("fil")
		}
	}
}
