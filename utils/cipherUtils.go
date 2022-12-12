package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"strings"
)

func Aes256Encode(content []byte, encryptionKey []byte) (encryptedContent []byte, IV []byte, err error) {
	bPlaintext := PKCS5Padding(content, aes.BlockSize)
	block, err := aes.NewCipher(encryptionKey)

	if err != nil {
		return nil, nil, err
	}

	IV, _ = GenerateRandomBytes(block.BlockSize())
	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, IV)
	mode.CryptBlocks(ciphertext, bPlaintext)

	return ciphertext, IV, err
}

func Aes256Decode(cipherText []byte, encryptionKey []byte, IV []byte) (decryptedContent []byte, err error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, IV)
	mode.CryptBlocks(cipherText, cipherText)

	cutTrailingSpaces := []byte(strings.TrimSpace(string(cipherText)))
	return cutTrailingSpaces, err
}

func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(cipherText, padText...)
}
