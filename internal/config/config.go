package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
	Seed     SeedConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Host           string
	Port           string
	ReadTimeout    int
	WriteTimeout   int
	AllowedOrigins []string
}

type AuthConfig struct {
	AccessSecret    string
	RefreshSecret   string
	EncodeIDSecret  string
	AccessTTL       time.Duration
	RefreshTTLShort time.Duration
	RefreshTTLLong  time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type AppConfig struct {
	Environment  string
	LogLevel     string
	CookieDomain string
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
			ReadTimeout:    getEnvAsInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout:   getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),
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
			Environment:  getEnv("ENVIRONMENT", "local"),
			LogLevel:     getEnv("LOG_LEVEL", "debug"),
			CookieDomain: getEnv("COOKIE_DOMAIN", ""),
		},
		Seed: SeedConfig{
			RootUser:     getEnv("ROOT_USER", ""),
			RootPassword: getEnv("ROOT_PASSWORD", ""),
		},
		Auth: AuthConfig{
			AccessSecret:    getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret:   getEnv("JWT_REFRESH_SECRET", ""),
			EncodeIDSecret:  getEnv("JWT_ENCODE_ID_SECRET", ""),
			AccessTTL:       time.Duration(getEnvAsInt("TTL_ACCESS", 600)) * time.Second,
			RefreshTTLShort: time.Duration(getEnvAsInt("TTL_REFRESH_SHORT", 86400)) * time.Second,
			RefreshTTLLong:  time.Duration(getEnvAsInt("TTL_REFRESH_LONG", 604800)) * time.Second,
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

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		return strings.Split(val, ",")
	}
	return defaultVal
}
