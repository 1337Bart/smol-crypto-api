package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"
	"github.com/1337Bart/smol-crypto-api/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/1337Bart/smol-crypto-api/internal/config"
	"github.com/1337Bart/smol-crypto-api/internal/server"
	"github.com/go-redis/redis/v8"

	internal_redis "github.com/1337Bart/smol-crypto-api/internal/repository/redis"
	_ "github.com/lib/pq"
)

func testOTLPConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to OTLP endpoint")
	return nil
}

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

	if err := testOTLPConnection(); err != nil {
		log.Printf("OTLP connection test failed: %v", err)
	}

	tp, err := server.InitTracer()
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	tracer := tp.Tracer("crypto-service")

	redisCache := internal_redis.NewCryptoCache(redisClient)
	postgresSQL := postgres.NewCryptoRepository(db, tracer)

	cryptoService := service.NewCryptoService(redisCache, postgresSQL, tracer)

	srv := server.New(cfg, cryptoService, tracer)

	ctx := context.Background()
	fmt.Println("Serving http and grpc ..")
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
