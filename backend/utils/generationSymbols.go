package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

// сессии для регистрации
func GenerateSessioID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func GenerationSessionCode(n int) (code string) {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code += string(chars[num.Int64()])
	}
	return
}
