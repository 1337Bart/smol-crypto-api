package main

import (
	"context"
	"database/sql"
	"fmt"
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
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	redisCache := internal_redis.NewCryptoCache(redisClient)

	// todo remove hardcode
	dbConfig := "host=localhost port=5432 user=crypto_app password=crypto_password dbname=crypto_db sslmode=disable"
	db, err := sql.Open("postgres", dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database")
	dbRepo := postgres.NewCryptoRepository(db)

	cryptoService := service.NewCryptoUpdateService(redisCache, dbRepo, []string{})

	ctx := context.Background()

	fmt.Println("Starting data fetch and save..")
	now := time.Now()
	cryptoService.UpdateCryptosSingle(ctx)
	//cryptoService.StartPeriodicUpdates(ctx)
	//select {}

	fmt.Printf("Data fetch and save took: %v\n", time.Since(now))
	fmt.Println("finished working, exiting..")
}
