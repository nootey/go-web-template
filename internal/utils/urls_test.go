package utils_test

import (
	"testing"

	"go-web-template/internal/config"
	"go-web-template/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestAPIBaseURL(t *testing.T) {
	local := &config.Config{}
	local.App.Environment = "local"
	local.Server.Host = "127.0.0.1"
	local.Server.Port = "8080"
	assert.Equal(t, "http://127.0.0.1:8080", utils.APIBaseURL(local))

	prod := &config.Config{}
	prod.App.Environment = "production"
	prod.Server.Host = "api.example.com"
	prod.Server.Port = "8080"
	assert.Equal(t, "https://api.example.com", utils.APIBaseURL(prod))
}

func TestWebClientBaseURL(t *testing.T) {
	local := &config.Config{}
	local.App.Environment = "local"
	local.WebClient.Domain = "localhost"
	local.WebClient.Port = "5173"
	assert.Equal(t, "http://localhost:5173", utils.WebClientBaseURL(local))

	prod := &config.Config{}
	prod.App.Environment = "production"
	prod.WebClient.Domain = "app.example.com"
	prod.WebClient.Port = "5173"
	assert.Equal(t, "https://app.example.com", utils.WebClientBaseURL(prod))
}
