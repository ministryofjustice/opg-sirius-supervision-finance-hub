package server

import (
	"os"
	"strconv"
)

type EnvironmentVars struct {
	Port                   string
	WebDir                 string
	SiriusURL              string
	SiriusPublicURL        string
	Prefix                 string
	BackendUrl             string
	SupervisionBillingTeam int
}

func NewEnvironmentVars() (EnvironmentVars, error) {
	supervisionBillingTeamId, err := strconv.Atoi(getEnv("SUPERVISION_BILLING_TEAM_ID", "41"))

	if err != nil {
		return EnvironmentVars{}, err
	}

	return EnvironmentVars{
		Port:                   getEnv("PORT", "1234"),
		WebDir:                 getEnv("WEB_DIR", "web"),
		SiriusURL:              getEnv("SIRIUS_URL", "http://host.docker.internal:8080"),
		SiriusPublicURL:        getEnv("SIRIUS_PUBLIC_URL", ""),
		Prefix:                 getEnv("PREFIX", ""),
		BackendUrl:             getEnv("BACKEND_URL", ""),
		SupervisionBillingTeam: supervisionBillingTeamId,
	}, nil
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}
