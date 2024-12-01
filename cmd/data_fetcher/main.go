package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/config"
	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"
	internal_redis "github.com/1337Bart/smol-crypto-api/internal/repository/redis"
	"github.com/1337Bart/smol-crypto-api/internal/service"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func main() {
	fmt.Println("Starting crypto data fetcher")
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
		log.Fatalf("Failed to connect to database: %v", err)
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

	ctx := context.Background()

	fmt.Println("Starting data fetch and save..")
	now := time.Now()
	cryptoService.UpdateCryptosSingle(ctx)
	//cryptoService.StartPeriodicUpdates(ctx)
	//select {}

	fmt.Printf("Data fetch and save took: %v\n", time.Since(now))
	fmt.Println("finished working, exiting..")
}
