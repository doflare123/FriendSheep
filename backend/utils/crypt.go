package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"

	"golang.org/x/crypto/argon2"
)

func generateSaltRaw(length int) []byte {
	salt := make([]byte, length)
	rand.Read(salt)
	return salt
}

func HashPassword(password string) (hashHex, saltHex string) {
	salt := generateSaltRaw(16)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return hex.EncodeToString(hash), hex.EncodeToString(salt)
}

func ComparePasswords(password, storedHash, saltHex string) bool {
	salt, _ := hex.DecodeString(saltHex)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	hashHex := hex.EncodeToString(hash)
	return subtle.ConstantTimeCompare([]byte(hashHex), []byte(storedHash)) == 1
}
