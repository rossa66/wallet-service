package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	ServerPort       int
	MaxDBConnections int
}

func LoadConfig() *Config {
	port, err := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
	if err != nil {
		log.Fatalf("Invalid POSTGRES_PORT: %v", err)
	}

	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		log.Fatalf("Invalid SERVER_PORT: %v", err)
	}

	maxConns, err := strconv.Atoi(getEnv("MAX_DB_CONNECTIONS", "50"))
	if err != nil {
		log.Fatalf("Invalid MAX_DB_CONNECTIONS: %v", err)
	}

	return &Config{
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     port,
		PostgresUser:     getEnv("POSTGRES_USER", "walletuser"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "walletpass"),
		PostgresDB:       getEnv("POSTGRES_DB", "walletdb"),
		ServerPort:       serverPort,
		MaxDBConnections: maxConns,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
