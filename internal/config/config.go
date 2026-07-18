package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Cors      CorsConfig      `mapstructure:"cors"`
	Database  DatabaseConfig  `mapstructure:"database"`
	App       AppConfig       `mapstructure:"app"`
	Seed      SeedConfig      `mapstructure:"seed"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Session   SessionConfig   `mapstructure:"session"`
	Mailer    MailerConfig    `mapstructure:"mailer"`
	Token     TokenConfig     `mapstructure:"token"`
	WebClient WebClientConfig `mapstructure:"web_client"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"name"`
}

type AppConfig struct {
	Environment  string `mapstructure:"environment"`
	LogLevel     string `mapstructure:"log_level"`
	CookieDomain string `mapstructure:"cookie_domain"`
}

type WebClientConfig struct {
	Domain string `mapstructure:"domain"`
	Port   string `mapstructure:"port"`
}

type SeedConfig struct {
	RootUser     string `mapstructure:"root_user"`
	RootPassword string `mapstructure:"root_password"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type SessionConfig struct {
	TTLHours           int `mapstructure:"ttl_hours"`
	RememberMeTTLHours int `mapstructure:"remember_me_ttl_hours"`
}

type MailerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
}

type TokenConfig struct {
	ConfirmTTLHours int `mapstructure:"confirm_ttl_hours"`
	ResetTTLMinutes int `mapstructure:"reset_ttl_minutes"`
}

var cfg *Config

// setDefaults doubles as the key registry: AutomaticEnv only resolves keys viper
// already knows about, so a key absent here is not overridable by env var.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", 10)
	v.SetDefault("server.write_timeout", 10)

	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3030", "http://localhost:5173"})

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "go-web-template")

	v.SetDefault("app.environment", "local")
	v.SetDefault("app.log_level", "debug")
	v.SetDefault("app.cookie_domain", "")

	v.SetDefault("web_client.domain", "localhost")
	v.SetDefault("web_client.port", "5173")

	v.SetDefault("seed.root_user", "root@local.host")
	v.SetDefault("seed.root_password", "password")

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("session.ttl_hours", 24)
	v.SetDefault("session.remember_me_ttl_hours", 720)

	v.SetDefault("mailer.host", "")
	v.SetDefault("mailer.port", 587)
	v.SetDefault("mailer.username", "")
	v.SetDefault("mailer.password", "")
	v.SetDefault("mailer.from", "no-reply@localhost")
	v.SetDefault("mailer.from_name", "Go Web Template")

	v.SetDefault("token.confirm_ttl_hours", 24)
	v.SetDefault("token.reset_ttl_minutes", 60)
}

func Load() error {
	v := viper.New()

	setDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	c := &Config{}
	if err := v.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := c.validate(); err != nil {
		return err
	}

	cfg = c
	return nil
}

// validate rejects config combinations that unmarshal fine but are unsafe to run with.
func (c *Config) validate() error {
	if c.App.Environment == "production" && c.Redis.Password == "" {
		return errors.New("redis password is required when app.environment is production")
	}
	return nil
}

func Get() *Config {
	if cfg == nil {
		panic("config not loaded")
	}
	return cfg
}
