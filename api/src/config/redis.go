package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	ctx         = context.Background()
)

func InicializarRedis() (*redis.Client, error) {
	// Verifica se a variável de ambiente REDIS_URL está definida
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("variável de ambiente REDIS_URL não definida")
	}

	// Cria o cliente Redis com timeouts
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         redisURL,
		DialTimeout:  10 * time.Second, // Timeout para estabelecer a conexão
		ReadTimeout:  30 * time.Second, // Timeout para operações de leitura
		WriteTimeout: 30 * time.Second, // Timeout para operações de escrita
	})

	// Testa a conexão com o Redis
	if err := TestarRedis(); err != nil {
		return nil, fmt.Errorf("falha ao conectar ao Redis: %v", err)
	}

	fmt.Println("Conectado ao Redis com sucesso!")
	return RedisClient, nil
}

// Testa a conexão com o Redis
func TestarRedis() error {
	if RedisClient == nil {
		return fmt.Errorf("cliente Redis não inicializado")
	}

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("erro ao testar conexão com Redis: %v", err)
	}

	fmt.Println("Conexão com Redis testada com sucesso!")
	return nil
}
