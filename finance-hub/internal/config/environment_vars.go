package config

import (
	"os"
	"strconv"
)

type EnvironmentVars struct {
	Port            string
	WebDir          string
	SiriusURL       string
	SiriusPublicURL string
	Prefix          string
	BackendURL      string
	JwtSecret       string
	JwtExpiry       int
}

func NewEnvironmentVars() (EnvironmentVars, error) {
	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY", "1"))
	return EnvironmentVars{
		Port:            getEnv("PORT", "1234"),
		WebDir:          getEnv("WEB_DIR", "web"),
		SiriusURL:       getEnv("SIRIUS_URL", "http://host.docker.internal:8080"),
		SiriusPublicURL: getEnv("SIRIUS_PUBLIC_URL", ""),
		Prefix:          getEnv("PREFIX", ""),
		BackendURL:      getEnv("BACKEND_URL", ""),
		JwtSecret:       getEnv("JWT_SECRET", "mysupersecrettestkeythatis128bits"),
		JwtExpiry:       jwtExpiry,
	}, nil
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}
