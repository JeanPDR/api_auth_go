package main

import (
	"auth-api/internal/auth"
	"auth-api/internal/database"
	"auth-api/internal/mailer"
	"fmt"
	"log"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Falha crítica ao conectar no banco: %v", err)
	}
	defer db.Close()

	mailSvc := mailer.NewMailer()

	authRepo := auth.NewRepository(db)
	authHandler := auth.NewHandler(authRepo, mailSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "API Online 2.0 🚀")
	})

	mux.HandleFunc("/register", authHandler.RegisterUser)
	mux.HandleFunc("/login", authHandler.LoginUser)
	mux.HandleFunc("/refresh", authHandler.RefreshToken)
	mux.HandleFunc("/verify", authHandler.VerifyEmail)
	mux.HandleFunc("/verify/resend", authHandler.ResendVerificationCode)
	mux.HandleFunc("/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("/reset-password", authHandler.ResetPassword)

	mux.HandleFunc("/dashboard", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Bem-vindo ao sistema protegido!", "seu_id": "%s"}`, userID)
	}))

	mux.HandleFunc("/me", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"authenticated": true}`))
	}))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:4200",
			"https://app.seudominio.com.br",
			"https://jeanpreis.com.br",
			"https://www.jeanpreis.com.br",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		Debug:            false,
	})

	handlerComCors := c.Handler(mux)

	fmt.Println("Servidor rodando na porta :8080 com CORS ativado!")
	if err := http.ListenAndServe(":8080", handlerComCors); err != nil {
		log.Fatal(err)
	}
}
