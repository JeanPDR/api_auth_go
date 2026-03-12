package auth

import (
	"auth-api/internal/mailer"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unicode"
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

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
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
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "JSON inválido ou contém campos não permitidos", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "E-mail inválido", http.StatusBadRequest)
		return
	}

	if !isValidPassword(req.Password) {
		http.Error(w, "A senha deve ter pelo menos 8 caracteres, incluindo uma letra maiúscula e um caractere especial.", http.StatusBadRequest)
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
	decoder.DisallowUnknownFields()
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

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Erro interno ao gerar refresh token", http.StatusInternalServerError)
		return
	}

	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.repo.SaveRefreshToken(r.Context(), user.ID, refreshToken, refreshExpiresAt); err != nil {
		http.Error(w, "Erro ao salvar sessão", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   7200,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login realizado com sucesso!",
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

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req ForgotPasswordRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	_, err := h.repo.GetUserByEmailForVerification(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Não existe nenhuma conta cadastrada com este e-mail.", http.StatusNotFound)
		return
	}

	code := GenerateRandomCode()
	expiresAt := time.Now().Add(30 * time.Minute)

	if err := h.repo.UpdateVerificationCode(r.Context(), req.Email, code, expiresAt); err != nil {
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	go func() {
		err := h.mailer.SendPasswordResetCode(req.Email, code)
		if err != nil {
			fmt.Printf("Erro silencioso ao enviar e-mail de reset: %v\n", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Código de recuperação enviado com sucesso para o seu e-mail.",
	})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req ResetPasswordRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if !isValidPassword(req.NewPassword) {
		http.Error(w, "A nova senha deve ter pelo menos 8 caracteres, incluindo uma letra maiúscula e um caractere especial.", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmailForVerification(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Dados incorretos", http.StatusBadRequest)
		return
	}

	if user.VerificationCode != req.Code || time.Now().After(user.ExpiresAt) {
		http.Error(w, "Código inválido ou expirado", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Erro interno de segurança", http.StatusInternalServerError)
		return
	}

	if err := h.repo.UpdatePassword(r.Context(), req.Email, hashedPassword); err != nil {
		http.Error(w, "Erro ao atualizar a senha", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Palavra-passe alterada com sucesso! Já pode fazer login.",
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Sessão inválida ou ausente", http.StatusUnauthorized)
		return
	}
	refreshTokenStr := cookie.Value

	user, err := h.repo.GetUserByRefreshToken(r.Context(), refreshTokenStr)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	if time.Now().After(user.RefreshExpiresAt) {
		http.Error(w, "Sessão expirada. Por favor, faça login novamente.", http.StatusUnauthorized)
		return
	}

	newToken, err := GenerateJWT(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Erro ao gerar novo token de acesso", http.StatusInternalServerError)
		return
	}

	newRefreshToken, _ := GenerateRefreshToken()
	newRefreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	_ = h.repo.SaveRefreshToken(r.Context(), user.ID, newRefreshToken, newRefreshExpiresAt)

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    newToken,
		Path:     "/",
		MaxAge:   7200,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Sessão renovada com sucesso!",
	})
}

func isValidPassword(s string) bool {
	if len(s) < 8 {
		return false
	}
	var hasUpper, hasSpecial bool
	for _, char := range s {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}
	return hasUpper && hasSpecial
}

func (h *Handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshTokenStr := cookie.Value
		user, err := h.repo.GetUserByRefreshToken(r.Context(), refreshTokenStr)
		if err == nil {
			_ = h.repo.SaveRefreshToken(r.Context(), user.ID, "", time.Now().Add(-1*time.Hour))
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logout realizado com sucesso!"}`))
}
