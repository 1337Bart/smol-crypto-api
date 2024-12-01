package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	GRPCPort string
	HTTPPort string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			GRPCPort: getEnvAsString("GRPC_PORT", "50051"),
			HTTPPort: getEnvAsString("HTTP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsString("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "crypto_app"),
			Password: getEnv("DB_PASSWORD", "crypto_password"),
			DBName:   getEnv("DB_NAME", "crypto_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsString("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
