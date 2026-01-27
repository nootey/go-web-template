package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
	Seed     SeedConfig
}

type ServerConfig struct {
	Host           string
	Port           string
	ReadTimeout    int
	WriteTimeout   int
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type AppConfig struct {
	Environment string
	LogLevel    string
}

type SeedConfig struct {
	RootUser     string
	RootPassword string
}

var cfg *Config

func Load() error {
	cfg = &Config{
		Server: ServerConfig{
			Host:           getEnv("SERVER_HOST", "127.0.0.1"),
			Port:           getEnv("SERVER_PORT", "8080"),
			ReadTimeout:    getEnvInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout:   getEnvInt("SERVER_WRITE_TIMEOUT", 10),
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "go-web-template"),
		},
		App: AppConfig{
			Environment: getEnv("ENVIRONMENT", "local"),
			LogLevel:    getEnv("LOG_LEVEL", "debug"),
		},
		Seed: SeedConfig{
			RootUser:     getEnv("ROOT_USER", ""),
			RootPassword: getEnv("ROOT_PASSWORD", ""),
		},
	}
	return nil
}

func Get() *Config {
	if cfg == nil {
		panic("config not loaded")
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvSlice(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		return strings.Split(val, ",")
	}
	return defaultVal
}
