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

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResendCodeRequest struct {
	Email string `json:"email"`
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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // 🚨 Rejeita JSON com campos extras!
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "JSON inválido ou contém campos não permitidos", http.StatusBadRequest)
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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // 🚨 Rejeita JSON com campos extras!
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "JSON inválido ou contém campos não permitidos", http.StatusBadRequest)
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

	if !user.IsVerified {
		http.Error(w, "Por favor, verifique o seu e-mail antes de fazer login.", http.StatusForbidden)
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

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() 
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "JSON inválido ou contém campos não permitidos", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmailForVerification(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	if user.IsVerified {
		http.Error(w, "Este e-mail já foi verificado anteriormente", http.StatusBadRequest)
		return
	}

	if user.VerificationCode != req.Code {
		http.Error(w, "Código de verificação incorreto", http.StatusUnauthorized)
		return
	}

	if time.Now().After(user.ExpiresAt) {
		http.Error(w, "O código de verificação expirou. Solicite um novo.", http.StatusUnauthorized)
		return
	}

	if err := h.repo.MarkUserAsVerified(r.Context(), req.Email); err != nil {
		http.Error(w, "Erro interno ao validar e-mail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "E-mail verificado com sucesso! Já pode fazer login.",
	})
}

func (h *Handler) ResendVerificationCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req ResendCodeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // 🚨 Rejeita JSON com campos extras!
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "JSON inválido ou contém campos não permitidos", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmailForVerification(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	if user.IsVerified {
		http.Error(w, "Este e-mail já foi verificado. Pode fazer login.", http.StatusBadRequest)
		return
	}

	newCode := GenerateRandomCode()
	expiresAt := time.Now().Add(2 * time.Hour)

	if err := h.repo.UpdateVerificationCode(r.Context(), req.Email, newCode, expiresAt); err != nil {
		http.Error(w, "Erro interno ao atualizar código", http.StatusInternalServerError)
		return
	}

	go func() {
		err := h.mailer.SendConfirmationCode(req.Email, newCode)
		if err != nil {
			fmt.Printf("Erro silencioso ao reenviar e-mail para %s: %v\n", req.Email, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Novo código de verificação enviado com sucesso!",
	})
}