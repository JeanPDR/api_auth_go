package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	os.Setenv("JWT_SECRET", "chave_super_secreta_para_testes")
	defer os.Unsetenv("JWT_SECRET")

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey).(string)
		if userID != "user-123" {
			t.Errorf("Esperava o ID 'user-123' no contexto, mas recebeu '%s'", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := AuthMiddleware(protectedHandler)

	validToken, err := GenerateJWT("user-123", "teste@exemplo.com")
	if err != nil {
		t.Fatalf("Erro inesperado ao gerar token de teste: %v", err)
	}

	tests := []struct {
		name           string
		cookieValue    string
		setCookie      bool
		expectedStatus int
	}{
		{
			name:           "Deve bloquear requisição sem cookie access_token",
			setCookie:      false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Deve bloquear token com formato inválido ou assinatura falsa",
			cookieValue:    validToken + "falso",
			setCookie:      true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Deve permitir o acesso com cookie válido",
			cookieValue:    validToken,
			setCookie:      true,
			expectedStatus: http.StatusOK, // 200 OK
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
			
			// 🚨 Agora o teste injeta o Cookie em vez do cabeçalho
			if tc.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: tc.cookieValue,
				})
			}

			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Cenário '%s': esperava status %v, mas recebeu %v", tc.name, tc.expectedStatus, status)
			}
		})
	}
}