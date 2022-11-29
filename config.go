package gogossip

import (
	"fmt"
)

type Config struct {
	FilterWithStorage string
	GossipNumber      int
	EncType           EncryptType
	Passphrase        string
}

func DefaultConfig() *Config {
	return &Config{
		GossipNumber: 2,
		EncType:      NON_SECURE_TYPE,
		Passphrase:   "",
	}
}

func (c *Config) validate() error {
	if c.EncType.String() == "" {
		return fmt.Errorf("invalid EncryptType %d", c.EncType)
	}
	if c.EncType != NON_SECURE_TYPE && c.Passphrase == "" {
		return fmt.Errorf("Passphrase required")
	}
	return nil
}
