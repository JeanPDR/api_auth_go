package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "minha_senha_super_segura"

	// 1. Testa a criação do Hash
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Erro inesperado ao gerar hash: %v", err)
	}
	if hash == password {
		t.Errorf("O hash não pode ser igual à senha original")
	}

	// 2. Testa a validação (Cenário de Sucesso)
	isValid := CheckPasswordHash(password, hash)
	if !isValid {
		t.Errorf("A senha correta deveria ser validada com o hash gerado")
	}

	// 3. Testa a validação (Cenário de Falha)
	isInvalid := CheckPasswordHash("senha_errada_123", hash)
	if isInvalid {
		t.Errorf("Uma senha incorreta não deveria ser validada")
	}
}