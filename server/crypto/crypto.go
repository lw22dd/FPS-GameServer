package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Crypto 提供数据传输加密解密功能，这里是对称加密功能
var (
	encryptionKey = []byte("32字节密钥1234567890123456")
)

func init() {
	// 确保密钥长度为32字节
	SetKey("32字节密钥1234567890123456")
}

func SetKey(key string) {
	if len(key) >= 32 {
		encryptionKey = []byte(key[:32])
	} else {
		paddedKey := make([]byte, 32)
		copy(paddedKey, []byte(key))
		encryptionKey = paddedKey
	}
}

func GetKey() []byte {
	return encryptionKey
}

// 加密逻辑
func Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("创建加密块失败: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("生成随机数失败: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密逻辑
func Decrypt(encryptedBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("创建解密块失败: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}

	nonce, ciphertextBytes := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	return string(plaintext), nil
}

func EncryptJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}
	return Encrypt(string(data))
}

func DecryptJSON(encryptedBase64 string, v interface{}) error {
	plaintext, err := Decrypt(encryptedBase64)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(plaintext), v)
}
