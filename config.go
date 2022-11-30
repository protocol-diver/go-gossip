package gogossip

import (
	"errors"
	"io/fs"
)

var (
	errInvalidGossipNumber = errors.New("invalid GossipNumber")
	errInvalidFilePath     = errors.New("invalid FilterWithStorage")
	errInvalidEncryptType  = errors.New("invalid EncryptType")
	errRequirePassphrase   = errors.New("required Passphrase")
)

type Config struct {
	// FilterWithStorage is the filter option variable. If the
	// value is nil, it means a memory filter. Set the path to
	// save the data if to use the storage filter.
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
	if c.GossipNumber < 2 {
		return errInvalidGossipNumber
	}
	if c.FilterWithStorage != "" {
		if !fs.ValidPath(c.FilterWithStorage) {
			return errInvalidFilePath
		}
	}
	if c.EncType.String() == "" {
		return errInvalidEncryptType
	}
	if c.EncType != NON_SECURE_TYPE && c.Passphrase == "" {
		return errRequirePassphrase
	}

	return nil
}
