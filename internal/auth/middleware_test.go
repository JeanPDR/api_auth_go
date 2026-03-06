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
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Deve bloquear requisição sem cabeçalho Authorization",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Deve bloquear token com formato inválido (sem Bearer)",
			authHeader:     "MeuTokenSecreto123",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Deve bloquear token com assinatura falsa",
			authHeader:     "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.falso.falso",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Deve permitir o acesso com token válido",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK, // 200 OK
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
			
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Cenário '%s': esperava status %v, mas recebeu %v", tc.name, tc.expectedStatus, status)
			}
		})
	}
}