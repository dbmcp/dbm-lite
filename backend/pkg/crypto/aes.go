package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

func padKey(key string) []byte {
	k := []byte(key)
	if len(k) >= 32 {
		return k[:32]
	}
	padded := make([]byte, 32)
	copy(padded, k)
	return padded
}

func EncryptAES(plainText string, keyStr string) (string, error) {
	key := padKey(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

func DecryptAES(cipherHex string, keyStr string) (string, error) {
	key := padKey(keyStr)
	cipherText, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("invalid cipher text")
	}
	nonce, cipherData := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}
