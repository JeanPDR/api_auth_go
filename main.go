package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Inicializa banco
	db := ConnectDB()
	defer db.Close()

	// Handler simples de teste
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "API Online 🚀")
	})

	fmt.Println("Servidor rodando na porta :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}