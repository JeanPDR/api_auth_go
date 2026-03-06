package main

import (
	"auth-api/internal/auth"
	"auth-api/internal/database"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Chamada correta recebendo os dois valores de retorno
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Falha crítica ao conectar no banco: %v", err)
	}
	defer db.Close()

	authRepo := auth.NewRepository(db)
	authHandler := auth.NewHandler(authRepo)

	// Rota de teste
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "API Online 🚀")
	})

	http.HandleFunc("/register", authHandler.RegisterUser)

	fmt.Println("Servidor rodando na porta :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}