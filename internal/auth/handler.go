package auth

import (
	"auth-api/internal/mailer"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	repo   *Repository
	mailer *mailer.Mailer
}

func NewHandler(repo *Repository, mailSvc *mailer.Mailer) *Handler {
	return &Handler{
		repo:   repo,
		mailer: mailSvc,
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Email == "" || len(req.Password) < 6 {
		http.Error(w, "E-mail inválido ou senha muito curta (mínimo de 6 caracteres)", http.StatusBadRequest)
		return
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Erro interno ao processar segurança", http.StatusInternalServerError)
		return
	}

	code := GenerateRandomCode()
	expiresAt := time.Now().Add(2 * time.Hour)

	user := &User{
		Email:            req.Email,
		PasswordHash:     hashedPassword,
		VerificationCode: code,
		ExpiresAt:        expiresAt,
	}

	if err := h.repo.Create(r.Context(), user); err != nil {
		http.Error(w, "Erro ao criar usuário. O e-mail já está em uso?", http.StatusConflict)
		return
	}

	go func() {
		err := h.mailer.SendConfirmationCode(user.Email, code)
		if err != nil {
			fmt.Printf("Erro silencioso ao enviar e-mail para %s: %v\n", user.Email, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário cadastrado com sucesso! Verifique seu e-mail.",
		"email":   req.Email,
	})
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateJWT(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Erro interno ao gerar credenciais de acesso", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login realizado com sucesso!",
		"token":   tokenString,
	})
}