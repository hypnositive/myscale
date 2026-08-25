package config

import (
	"os"
)

type Config struct {
	GRPCPort string
	DBDSN    string
}

func Load() *Config {
	return &Config{
		GRPCPort: getEnv("GRPC_PORT", "50051"),
		DBDSN:    getEnv("DB_URL", "postgres://postgres:secret@localhost:5432/mydatabase?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}