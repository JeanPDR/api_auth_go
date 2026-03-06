package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth-api/internal/database"

	"github.com/joho/godotenv"
)

func TestRegisterUser(t *testing.T) {
	
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Logf("Aviso: não foi possível carregar o ficheiro .env (pode ignorar se estiver no CI/CD)")
	}

	
	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Falha ao ligar à base de dados no teste: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	handler := NewHandler(repo)

	// --- CENÁRIO 1: Falha por Senha Curta ---
	t.Run("Deve retornar 400 se a senha for curta", func(t *testing.T) {
		payload := []byte(`{"email":"teste@invalido.com", "password":"123"}`)
		
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Código de status incorreto: obteve %v, esperava %v", status, http.StatusBadRequest)
		}
	})

	// --- CENÁRIO 2: Sucesso no Registo ---
	t.Run("Deve retornar 201 ao cadastrar utilizador válido", func(t *testing.T) {
		emailUnico := fmt.Sprintf("qa_teste_%d@exemplo.com", time.Now().Unix())
		payload := []byte(fmt.Sprintf(`{"email":"%s", "password":"senha_forte_123"}`, emailUnico))
		
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("Código de status incorreto: obteve %v, esperava %v. Resposta: %s", status, http.StatusCreated, rr.Body.String())
		}
	})
}