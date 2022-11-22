package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
)

type AES256_CBC struct{}

func (a AES256_CBC) Encrypt(key string, plain []byte) ([]byte, error) {
	var (
		kb = sha256.Sum256([]byte(key))
		iv = kb[:aes.BlockSize]
	)

	block, err := aes.NewCipher(kb[:])
	if err != nil {
		return nil, err
	}

	padded := padPKCS7(plain, block.BlockSize())

	enc := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(padded))

	enc.CryptBlocks(cipherText, padded)

	return cipherText, nil
}

func (a AES256_CBC) Decrypt(key string, data []byte) ([]byte, error) {
	var (
		kb = sha256.Sum256([]byte(key))
		iv = kb[:aes.BlockSize]
	)

	block, err := aes.NewCipher(kb[:])
	if err != nil {
		return nil, err
	}

	dec := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(data))

	dec.CryptBlocks(plainText, data)

	trimed := trimPKCS5(plainText)
	if trimed == nil {
		return nil, errors.New("invalid passphrase")
	}

	return trimed, nil
}

func padPKCS7(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, padText...)
}

func trimPKCS5(text []byte) []byte {
	padding := text[len(text)-1]
	idx := len(text) - int(padding)
	if idx < 0 {
		return nil
	}
	return text[:idx]
}
