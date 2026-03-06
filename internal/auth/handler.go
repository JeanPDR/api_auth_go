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

// Handler agrupa as dependências (como o banco de dados)
type Handler struct {
	repo *Repository
}

func NewHandler (repo *Repository) * Handler{
	return &Handler{repo: repo}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// 1. Apenas aceita requisições POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// 2. Lê o corpo da requisição e transforma na struct RegisterRequest
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3. Validação básica (evita quebrar o banco)
	if req.Email == "" || len(req.Password) < 6 {
		http.Error(w, "E-mail inválido ou senha muito curta (mínimo 6 caracteres)", http.StatusBadRequest)
		return
	}

	// 4. Protege a senha com Argon2id
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Erro interno ao processar segurança", http.StatusInternalServerError)
		return
	}

	// 5. Gera o código de 6 dígitos e define que ele expira em 2 horas
	code := GenerateRandomCode()
	expiresAt := time.Now().Add(2 * time.Hour)

	// 6. Monta o modelo de Usuário e tenta salvar no banco
	user := &User{
		Email:            req.Email,
		PasswordHash:     hashedPassword,
		VerificationCode: code,
		ExpiresAt:        expiresAt,
	}

	if err := h.repo.Create(r.Context(), user); err != nil {
		// Se der erro aqui, muito provavelmente o e-mail já existe (UNIQUE constraint)
		http.Error(w, "Erro ao criar usuário. O e-mail já está em uso?", http.StatusConflict)
		return
	}

	// 7. Retorna sucesso! (O disparo de e-mail entrará aqui depois)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário cadastrado com sucesso! Verifique seu e-mail.",
		"email":   req.Email,
	})
}