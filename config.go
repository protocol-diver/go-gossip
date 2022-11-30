package gogossip

import (
	"fmt"
)

type Config struct {
	// If the value is nil, it means a memory filter. Set the
	// path to save the data if to use the storage filter,
	FilterWithStorage string

	// GossipNumber means the number of peers to make pull
	// requests per pullInterval. This number must be greater
	// than 2, and if set to greater than the total number of
	// existing peers, it means broadcasting.
	GossipNumber int

	EncType    EncryptType
	Passphrase string
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
