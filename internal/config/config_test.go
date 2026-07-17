package config

import "testing"

// Guards the defaults-as-registry contract: a key must be in setDefaults to be
// overridable by env var, and the env name is the key with dots replaced by
// underscores.
func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_HOST", "db")
	t.Setenv("DATABASE_NAME", "go_template")
	t.Setenv("APP_ENVIRONMENT", "production")
	t.Setenv("SEED_ROOT_USER", "seeded@example.com")
	t.Setenv("REDIS_HOST", "redis")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("SESSION_TTL_HOURS", "48")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example,https://b.example")

	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	c := Get()

	if c.Database.Host != "db" {
		t.Errorf("Database.Host = %q, want %q", c.Database.Host, "db")
	}
	if c.Database.DBName != "go_template" {
		t.Errorf("Database.DBName = %q, want %q", c.Database.DBName, "go_template")
	}
	if c.App.Environment != "production" {
		t.Errorf("App.Environment = %q, want %q", c.App.Environment, "production")
	}
	if c.Seed.RootUser != "seeded@example.com" {
		t.Errorf("Seed.RootUser = %q, want %q", c.Seed.RootUser, "seeded@example.com")
	}
	if c.Redis.Host != "redis" || c.Redis.Port != 6380 {
		t.Errorf("Redis = %+v, want host redis port 6380", c.Redis)
	}
	if c.Session.TTLHours != 48 {
		t.Errorf("Session.TTLHours = %d, want 48", c.Session.TTLHours)
	}
	want := []string{"https://a.example", "https://b.example"}
	if len(c.Cors.AllowedOrigins) != len(want) {
		t.Fatalf("Cors.AllowedOrigins = %#v, want %#v", c.Cors.AllowedOrigins, want)
	}
	for i, o := range want {
		if c.Cors.AllowedOrigins[i] != o {
			t.Errorf("Cors.AllowedOrigins[%d] = %q, want %q", i, c.Cors.AllowedOrigins[i], o)
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	c := Get()

	if c.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want %q", c.Server.Port, "8080")
	}
	if c.Redis.Port != 6379 || c.Redis.DB != 0 {
		t.Errorf("Redis = %+v, want port 6379 db 0", c.Redis)
	}
	if c.Session.TTLHours != 24 {
		t.Errorf("Session.TTLHours = %d, want 24", c.Session.TTLHours)
	}
	if c.Session.RememberMeTTLHours != 720 {
		t.Errorf("Session.RememberMeTTLHours = %d, want 720", c.Session.RememberMeTTLHours)
	}
}
