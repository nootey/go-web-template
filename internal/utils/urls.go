package utils

import "go-web-template/internal/config"

// APIBaseURL returns the API base URL; https with no port in production.
func APIBaseURL(cfg *config.Config) string {
	return buildBaseURL(cfg.App.Environment == "production", cfg.Server.Host, cfg.Server.Port)
}

// WebClientBaseURL returns the web client base URL; https with no port in production.
func WebClientBaseURL(cfg *config.Config) string {
	return buildBaseURL(cfg.App.Environment == "production", cfg.WebClient.Domain, cfg.WebClient.Port)
}

func buildBaseURL(production bool, host, port string) string {
	scheme := "http://"
	if production {
		scheme = "https://"
		port = ""
	}

	base := scheme + host
	if port != "" {
		base += ":" + port
	}
	return base
}
