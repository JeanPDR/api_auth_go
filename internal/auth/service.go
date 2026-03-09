package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	return hash, err
}

func CheckPasswordHash(password, hash string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false
	}
	return match
}

func GenerateRandomCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Add(n, big.NewInt(100000)))
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Usamos URLEncoding para que seja seguro trafegar via HTTP
	return base64.URLEncoding.EncodeToString(b), nil
}