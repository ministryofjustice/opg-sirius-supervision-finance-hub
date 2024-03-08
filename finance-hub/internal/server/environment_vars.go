package server

import (
	"os"
)

type EnvironmentVars struct {
	Port            string
	WebDir          string
	SiriusURL       string
	SiriusPublicURL string
	Prefix          string
	BackendUrl      string
}

func NewEnvironmentVars() (EnvironmentVars, error) {

	return EnvironmentVars{
		Port:            getEnv("PORT", "1234"),
		WebDir:          getEnv("WEB_DIR", "web"),
		SiriusURL:       getEnv("SIRIUS_URL", "http://host.docker.internal:8080"),
		SiriusPublicURL: getEnv("SIRIUS_PUBLIC_URL", ""),
		Prefix:          getEnv("PREFIX", ""),
		BackendUrl:      getEnv("BACKEND_URL", ""),
	}, nil
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}
