package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// ConnectDB inicializa a conexão e retorna o pool e um possível erro
func ConnectDB() (*pgxpool.Pool, error) {
	// Carrega o .env se existir
	_ = godotenv.Load()

	// Configuração manual para evitar erros com caracteres especiais na senha
	config, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}

	config.ConnConfig.User = os.Getenv("DB_USER")
	config.ConnConfig.Password = os.Getenv("DB_PASSWORD")
	config.ConnConfig.Host = os.Getenv("DB_HOST")
	config.ConnConfig.Database = os.Getenv("DB_NAME")
	config.ConnConfig.Port = 5432

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	// Teste de conexão
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	fmt.Println("✅ Banco de dados conectado!")
	return pool, nil
}