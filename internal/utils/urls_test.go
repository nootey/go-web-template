package utils_test

import (
	"go-web-template/config"
	"go-web-template/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
