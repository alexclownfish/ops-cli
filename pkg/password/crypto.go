package password

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func Encrypt(plaintext, key string) (string, error) {
	// 尝试Base64解码密钥
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(keyBytes) != 32 {
		// 如果解码失败或长度不对，直接使用原始字符串
		keyBytes = []byte(key)
	}
	if len(keyBytes) != 32 {
		return "", fmt.Errorf("密钥长度必须为32字节")
	}
	
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
func Decrypt(ciphertext, key string) (string, error) {
	// 尝试Base64解码密钥
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(keyBytes) != 32 {
		// 如果解码失败或长度不对，直接使用原始字符串
		keyBytes = []byte(key)
	}
	if len(keyBytes) != 32 {
		return "", fmt.Errorf("密钥长度必须为32字节")
	}
	
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文数据太短")
	}
	
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}

func GenerateMasterKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
