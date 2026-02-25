package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	errEmptyFilePath = errors.New("file path is empty")
)

// LoadPublicKey загружает публичный ключ из файла
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	if path == "" {
		return nil, errEmptyFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA public key")
	}

	return rsaPub, nil
}

// LoadPrivateKey загружает приватный ключ из файла
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return nil, errEmptyFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Попробуем PKCS1 формат
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		return priv.(*rsa.PrivateKey), nil
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA private key")
	}

	return rsaPriv, nil
}

// Encrypt шифрует данные с помощью публичного ключа RSA
// Автоматически выбирает подходящий алгоритм в зависимости от размера данных
func Encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return data, nil
	}

	keySize := publicKey.Size()
	maxChunkSize := keySize - 42 // OAEP padding overhead

	if len(data) <= maxChunkSize {
		return encryptSmallData(data, publicKey)
	}

	return encryptLargeData(data, publicKey)
}

// encryptSmallData шифрует маленькие данные напрямую с помощью RSA-OAEP
func encryptSmallData(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(nil, rand.Reader, publicKey, data, nil)
}

// encryptLargeData шифрует большие данные используя гибридное шифрование:
// данные шифруются AES-GCM, а ключ AES шифруется RSA-OAEP
func encryptLargeData(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	// Генерируем случайный AES ключ
	aesKey := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Шифруем данные AES ключом
	encryptedData, err := encryptAES(data, aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with AES: %w", err)
	}

	// Шифруем AES ключ публичным RSA ключом
	encryptedKey, err := rsa.EncryptOAEP(nil, rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	// Объединяем зашифрованный ключ и данные
	return combineEncryptedKeyAndData(encryptedKey, encryptedData), nil
}

// combineEncryptedKeyAndData объединяет зашифрованный AES ключ и данные в один массив
// Формат: [4 байта длины ключа][зашифрованный ключ][зашифрованные данные]
func combineEncryptedKeyAndData(encryptedKey, encryptedData []byte) []byte {
	result := make([]byte, 4+len(encryptedKey)+len(encryptedData))
	// Первые 4 байта - длина зашифрованного ключа
	result[0] = byte(len(encryptedKey) >> 24)
	result[1] = byte(len(encryptedKey) >> 16)
	result[2] = byte(len(encryptedKey) >> 8)
	result[3] = byte(len(encryptedKey))
	copy(result[4:], encryptedKey)
	copy(result[4+len(encryptedKey):], encryptedData)
	return result
}

// Decrypt расшифровывает данные с помощью приватного ключа RSA
func Decrypt(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return encryptedData, nil
	}

	keySize := privateKey.Size()

	// Проверяем, является ли это гибридным шифрованием (первые 4 байта - длина ключа)
	if len(encryptedData) > 4 {
		keyLen := int(encryptedData[0])<<24 | int(encryptedData[1])<<16 | int(encryptedData[2])<<8 | int(encryptedData[3])

		// Если длина ключа разумная (меньше размера RSA ключа), это гибридное шифрование
		if keyLen > 0 && keyLen < keySize && len(encryptedData) > 4+keyLen {
			// Извлекаем зашифрованный AES ключ
			encryptedKey := encryptedData[4 : 4+keyLen]
			encryptedPayload := encryptedData[4+keyLen:]

			// Расшифровываем AES ключ
			aesKey, err := rsa.DecryptOAEP(nil, rand.Reader, privateKey, encryptedKey, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
			}

			// Расшифровываем данные
			return decryptAES(encryptedPayload, aesKey)
		}
	}

	// Прямое RSA шифрование
	return rsa.DecryptOAEP(nil, rand.Reader, privateKey, encryptedData, nil)
}

// encryptAES шифрует данные с помощью AES-GCM
func encryptAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Создаем GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Генерируем nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Шифруем данные
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decryptAES расшифровывает данные с помощью AES-GCM
func decryptAES(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Создаем GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Извлекаем nonce
	nonceSize := aesGCM.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Расшифровываем данные
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// DeriveKey создает ключ из пароля с помощью PBKDF2
func DeriveKey(password []byte, salt []byte) []byte {
	hash := sha256.Sum256(append(password, salt...))
	return hash[:]
}
