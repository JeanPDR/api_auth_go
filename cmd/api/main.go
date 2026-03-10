package main

import (
	"auth-api/internal/auth"
	"auth-api/internal/database"
	"auth-api/internal/mailer"
	"fmt"
	"log"
	"net/http"
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

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "API Online 2.0 🚀")
	})

	http.HandleFunc("/register", authHandler.RegisterUser)
	http.HandleFunc("/login", authHandler.LoginUser)
	http.HandleFunc("/refresh", authHandler.RefreshToken)
	http.HandleFunc("/verify", authHandler.VerifyEmail)
	http.HandleFunc("/verify/resend", authHandler.ResendVerificationCode)
	http.HandleFunc("/forgot-password", authHandler.ForgotPassword)
	http.HandleFunc("/reset-password", authHandler.ResetPassword)

	http.HandleFunc("/dashboard", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Bem-vindo ao sistema protegido!", "seu_id": "%s"}`, userID)
	}))

	fmt.Println("Servidor rodando na porta :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}