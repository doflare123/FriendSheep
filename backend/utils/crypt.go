package utils

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

func generateSalt() string {
	salt := make([]byte, 16)
	rand.Read(salt)
	return hex.EncodeToString([]byte(salt))
}

func HashPassword(password string) (hashHex, Salt string) {
	Salt = generateSalt()
	combo := []byte(password + Salt)
	hash := blake2b.Sum256(combo)
	hashHex = hex.EncodeToString(hash[:])
	return
}

func CompareSalts(salt1, salt2 string) bool {
	return salt1 == salt2
}
