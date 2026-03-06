package auth

import (
	"encoding/json"
	"net/http"
	"time"
)

type RegisterRequest struct{
	Email string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	repo *Repository
}

func NewHandler (repo *Repository) * Handler{
	return &Handler{repo: repo}
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
		http.Error(w, "E-mail inválido ou senha muito curta (mínimo 6 caracteres)", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário cadastrado com sucesso! Verifique seu e-mail.",
		"email":   req.Email,
	})
}