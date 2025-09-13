package hasher

import (
	"crypto/sha256"
	"encoding/hex"
)

func Hash(data, key string) string {
	hash := sha256.New()
	hash.Write([]byte(data + key))
	return hex.EncodeToString(hash.Sum(nil))
}
func VerifyHash(data, key, hash string) bool {
	return Hash(data, key) == hash
}
