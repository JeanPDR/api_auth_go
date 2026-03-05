package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func ConnectDB() *pgxpool.Pool {
    err := godotenv.Load()
    if err != nil {
        log.Println("Aviso: Arquivo .env não encontrado")
    }

    // Criamos um objeto de configuração vazio
    config, err := pgxpool.ParseConfig("") 
    if err != nil {
        log.Fatalf("Erro ao iniciar config: %v", err)
    }

    // Atribuímos os valores individualmente
    config.ConnConfig.User = os.Getenv("DB_USER")
    config.ConnConfig.Password = os.Getenv("DB_PASSWORD")
    config.ConnConfig.Host = os.Getenv("DB_HOST")
    config.ConnConfig.Port = 5432 // ou strconv.Atoi(os.Getenv("DB_PORT"))
    config.ConnConfig.Database = os.Getenv("DB_NAME")

    // Cria o pool com a config limpa
    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        log.Fatalf("Erro ao conectar ao banco: %v", err)
    }

    if err := pool.Ping(context.Background()); err != nil {
        log.Fatalf("Banco inacessível: %v", err)
    }

    fmt.Println("✅ Conectado com sucesso (mesmo com caracteres especiais)!")
    return pool
}