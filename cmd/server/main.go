package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/client/coingecko"
	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"

	"github.com/1337Bart/smol-crypto-api/internal/service"
	"log"
	"time"

	"github.com/1337Bart/smol-crypto-api/internal/config"
	"github.com/1337Bart/smol-crypto-api/internal/server"
	"github.com/go-redis/redis/v8"

	redis_cache "github.com/1337Bart/smol-crypto-api/internal/repository/redis"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	coinGeckoClient := coingecko.NewClient(1 * time.Minute)

	postgresDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	postgresSQL := postgres.NewCryptoRepository(db)

	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	redisCache := redis_cache.NewCryptoCache(redisClient)

	cryptoService := service.NewCryptoService(coinGeckoClient, postgresSQL, &redisCache)

	srv := server.New(cfg, cryptoService)

	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
