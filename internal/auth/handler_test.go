package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth-api/internal/database"
	"auth-api/internal/mailer"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, *Repository) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Logf("Aviso: não foi possível carregar o ficheiro .env (pode ignorar no CI)")
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Falha ao ligar à base de dados no teste: %v", err)
	}

	_, err = db.Exec(context.Background(), `
		DROP TABLE IF EXISTS users;
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			is_verified BOOLEAN DEFAULT FALSE,
			verification_code VARCHAR(10),
			verification_expires_at TIMESTAMP,
			refresh_token TEXT,
			refresh_expires_at TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Falha ao criar tabela de testes: %v", err)
	}

	repo := NewRepository(db)
	return db, repo 
}

func TestRegisterUser(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close() 

	dummyMailer := &mailer.Mailer{}
	handler := NewHandler(repo, dummyMailer)

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

	t.Run("Deve retornar 201 ao cadastrar utilizador válido", func(t *testing.T) {
		payload := []byte(`{"email":"novo_teste@exemplo.com", "password":"senha_forte_123"}`)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("Código de status incorreto: obteve %v, esperava %v. Resposta: %s", status, http.StatusCreated, rr.Body.String())
		}
	})
}

func TestLoginUser(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close() 

	dummyMailer := &mailer.Mailer{}
	handler := NewHandler(repo, dummyMailer)

	// Inserir um utilizador manualmente para o teste
	hashedPassword, _ := HashPassword("senha_valida_123")
	userNaoVerificado := &User{
		Email:            "bloqueado@exemplo.com",
		PasswordHash:     hashedPassword,
		VerificationCode: "123456",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
		IsVerified:       false, 
	}
	_ = repo.Create(context.Background(), userNaoVerificado)

	t.Run("Deve bloquear login (403) se o e-mail não estiver verificado", func(t *testing.T) {
		payload := []byte(`{"email":"bloqueado@exemplo.com", "password":"senha_valida_123"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.LoginUser(rr, req)

		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("Esperava bloqueio 403, mas obteve %v. Resposta: %s", status, rr.Body.String())
		}
	})
}