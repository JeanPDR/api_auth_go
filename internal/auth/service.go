package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/alexedwards/argon2id"
)

// HashPassword transforma a senha pura em um hash seguro usando Argon2id
func HashPassword(password string) (string, error) {
	// O DefaultParams usa as configurações recomendadas atuais de memória e iterações.
	// Ele gera o salt automaticamente e retorna a string no formato PHC.
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	return hash, err
}

// CheckPasswordHash compara a senha digitada com o hash Argon2id do banco
func CheckPasswordHash(password, hash string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false
	}
	return match
}

// GenerateRandomCode cria um código de 6 dígitos para e-mail
func GenerateRandomCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Add(n, big.NewInt(100000)))
}