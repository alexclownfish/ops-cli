package password

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// 使用Base64编码的32字节密钥
	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="

	tests := []struct {
		name      string
		plaintext string
	}{
		{"简单密码", "password123"},
		{"复杂密码", "Abc123#xyz~test"},
		{"中文密码", "密码测试123"},
		{"空密码", ""},
		{"长密码", "this_is_a_very_long_password_that_should_still_work_correctly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.plaintext, masterKey)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// 确保加密后的内容与原文不同（空字符串除外）
			if tt.plaintext != "" && encrypted == tt.plaintext {
				t.Errorf("加密后内容与原文相同")
			}

			decrypted, err := Decrypt(encrypted, masterKey)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	// 同一明文加密多次，应产生不同密文（因为随机nonce）
	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="
	plaintext := "password123"

	encrypted1, err := Encrypt(plaintext, masterKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	encrypted2, err := Encrypt(plaintext, masterKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Errorf("两次加密产生相同密文，可能nonce没有随机化")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	correctKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="
	wrongKey := "YW5vdGhlciAzMiBieXRlIGtleSBmb3IgdGVzdGluZyE="
	plaintext := "password123"

	encrypted, err := Encrypt(plaintext, correctKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = Decrypt(encrypted, wrongKey)
	if err == nil {
		t.Errorf("使用错误密钥解密成功，应该失败")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="

	tests := []struct {
		name       string
		ciphertext string
	}{
		{"无效Base64", "not-valid-base64!!!"},
		{"空字符串", ""},
		{"太短", "YWJj"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext, masterKey)
			if err == nil {
				t.Errorf("解密无效密文成功，应该失败")
			}
		})
	}
}

func TestGenerateMasterKey(t *testing.T) {
	key1, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey() error = %v", err)
	}

	key2, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey() error = %v", err)
	}

	// 确保两次生成的密钥不同
	if key1 == key2 {
		t.Errorf("两次生成的密钥相同")
	}

	// 确保密钥长度正确（Base64编码的32字节）
	if len(key1) < 40 {
		t.Errorf("密钥长度 = %v, too short", len(key1))
	}
}

func TestEncryptWithInvalidKeyLength(t *testing.T) {
	plaintext := "password123"

	tests := []struct {
		name string
		key  string
	}{
		{"密钥太短", "short"},
		{"空密钥", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Encrypt(plaintext, tt.key)
			if err == nil {
				t.Errorf("使用无效密钥加密成功，应该失败")
			}
		})
	}
}
