package gogossip

import (
	"errors"
	"testing"
)

func TestConfig(t *testing.T) {
	tds := []struct {
		cfg *Config
		err error
	}{
		{
			cfg: &Config{
				FilterWithStorage: "",
				GossipNumber:      0,
				EncType:           NON_SECURE_TYPE,
				Passphrase:        "",
			},
			err: errInvalidGossipNumber,
		},
		{
			cfg: &Config{
				FilterWithStorage: "//:",
				GossipNumber:      2,
				EncType:           NON_SECURE_TYPE,
				Passphrase:        "",
			},
			err: errInvalidFilePath,
		},
		{
			cfg: &Config{
				FilterWithStorage: "",
				GossipNumber:      2,
				EncType:           AES256_CBC_TYPE,
				Passphrase:        "",
			},
			err: errRequirePassphrase,
		},
		{
			cfg: &Config{
				FilterWithStorage: "data",
				GossipNumber:      2,
				EncType:           AES256_CBC_TYPE,
				Passphrase:        "1234",
			},
			err: nil,
		},
		{
			cfg: &Config{
				FilterWithStorage: "",
				GossipNumber:      2,
				EncType:           NON_SECURE_TYPE,
				Passphrase:        "1234",
			},
			err: nil,
		},
		{
			cfg: DefaultConfig(),
			err: nil,
		},
	}

	for _, td := range tds {
		err := td.cfg.validate()
		if !errors.Is(err, td.err) {
			t.Fatalf("TestConfig failure, want: %v, got: %v", td.err, err)
		}
	}

}
