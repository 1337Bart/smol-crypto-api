package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"

	"github.com/1337Bart/smol-crypto-api/internal/service"
	"log"

	"github.com/1337Bart/smol-crypto-api/internal/config"
	"github.com/1337Bart/smol-crypto-api/internal/server"
	"github.com/go-redis/redis/v8"

	internal_redis "github.com/1337Bart/smol-crypto-api/internal/repository/redis"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("stating server..")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	dbConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)
	db, err := sql.Open("postgres", dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Connected to database")

	redisCache := internal_redis.NewCryptoCache(redisClient)
	postgresSQL := postgres.NewCryptoRepository(db)

	cryptoService := service.NewCryptoService(redisCache, postgresSQL)

	srv := server.New(cfg, cryptoService)

	ctx := context.Background()
	fmt.Println("Serving http and grpc ..")
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
