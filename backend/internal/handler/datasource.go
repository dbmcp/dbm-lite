package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func currentUser(c *gin.Context) string {
	v, _ := c.Get("username")
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func obfuscate(pwd string) string {
	if pwd == "" {
		return ""
	}
	key := []byte("dbm-lite-secure-key-xxxx")[:32]
	block, err := aes.NewCipher(key)
	if err != nil {
		return base64.StdEncoding.EncodeToString([]byte(pwd))
	}
	iv := []byte("dbm-lite-iv-x")[:aes.BlockSize]
	src := []byte(pwd)
	padded := pkcs7Pad(src, block.BlockSize())
	mode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(padded))
	mode.CryptBlocks(out, padded)
	return base64.StdEncoding.EncodeToString(out)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

func generateDatasourceId() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func parseInt(s string, defaultValue int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultValue
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultValue
}