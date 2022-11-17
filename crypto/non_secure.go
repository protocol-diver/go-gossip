package crypto

type NON_SECURE struct{}

func (n NON_SECURE) Encrypt(passphrase string, buf []byte) ([]byte, error) {
	return buf, nil
}

func (n NON_SECURE) Decrypt(passphrase string, buf []byte) ([]byte, error) {
	return buf, nil
}
