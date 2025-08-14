package main

import (
	"context"
	"errors"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/paginate"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/server"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

type Envs struct {
	webDir           string
	siriusURL        string
	siriusPublicURL  string
	backendURL       string
	prefix           string
	port             string
	jwtSecret        string
	billingTeamID    int
	showDirectDebits bool
	allpayHost       string
	allpayAPIKey     string
	allpaySchemeCode string
	holidayAPIURL    string
}

func parseEnvs() (*Envs, error) {
	envs := map[string]string{
		"SIRIUS_URL":                  os.Getenv("SIRIUS_URL"),
		"SIRIUS_PUBLIC_URL":           os.Getenv("SIRIUS_PUBLIC_URL"),
		"PREFIX":                      os.Getenv("PREFIX"),
		"BACKEND_URL":                 os.Getenv("BACKEND_URL"),
		"SUPERVISION_BILLING_TEAM_ID": os.Getenv("SUPERVISION_BILLING_TEAM_ID"),
		"PORT":                        os.Getenv("PORT"),
		"JWT_SECRET":                  os.Getenv("JWT_SECRET"),
	}

	var missing []error
	for k, v := range envs {
		if v == "" {
			missing = append(missing, errors.New("missing environment variable: "+k))
		}
	}

	billingTeamId, err := strconv.Atoi(envs["SUPERVISION_BILLING_TEAM_ID"])
	if err != nil {
		missing = append(missing, errors.New("invalid SUPERVISION_BILLING_TEAM_ID"))
	}

	if len(missing) > 0 {
		return nil, errors.Join(missing...)
	}

	return &Envs{
		siriusURL:        envs["SIRIUS_URL"],
		siriusPublicURL:  envs["SIRIUS_PUBLIC_URL"],
		prefix:           envs["PREFIX"],
		backendURL:       envs["BACKEND_URL"],
		jwtSecret:        envs["JWT_SECRET"],
		billingTeamID:    billingTeamId,
		webDir:           "web",
		port:             envs["PORT"],
		showDirectDebits: os.Getenv("SHOW_DIRECT_DEBITS") == "1",
		allpayHost:       os.Getenv("ALLPAY_HOST"), // TODO: Move these to checked values once Direct Debits is ready for production
		allpayAPIKey:     os.Getenv("ALLPAY_API_KEY"),
		allpaySchemeCode: "OPGB",
		holidayAPIURL:    os.Getenv("HOLIDAY_API_URL"),
	}, nil
}

func main() {
	ctx := context.Background()
	logger := telemetry.NewLogger("opg-sirius-supervision-finance-hub")

	err := run(ctx, logger)
	if err != nil {
		logger.Error("fatal startup error", slog.Any("err", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	exportTraces := env.Get("TRACING_ENABLED", "0") == "1"

	shutdown, err := telemetry.StartTracerProvider(ctx, logger, exportTraces)
	defer shutdown()
	if err != nil {
		return err
	}

	envs, err := parseEnvs()
	if err != nil {
		return err
	}

	alllpayClient := allpay.NewClient(http.DefaultClient, envs.allpayHost, envs.allpayAPIKey, envs.allpaySchemeCode)

	client := api.NewClient(
		http.DefaultClient,
		&auth.JWT{
			Secret: envs.jwtSecret,
		},
		api.Envs{
			SiriusURL:     envs.siriusURL,
			BackendURL:    envs.backendURL,
			HolidayAPIURL: envs.holidayAPIURL,
		},
		alllpayClient)

	templates := createTemplates(envs)

	s := &http.Server{
		Addr: ":" + envs.port,
		Handler: server.New(logger, client, templates, server.Envs{
			Port:             envs.port,
			WebDir:           envs.webDir,
			SiriusURL:        envs.siriusURL,
			SiriusPublicURL:  envs.siriusPublicURL,
			Prefix:           envs.prefix,
			BackendURL:       envs.backendURL,
			BillingTeamID:    envs.billingTeamID,
			ShowDirectDebits: envs.showDirectDebits,
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			logger.Error("listen and server error", slog.Any("err", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("Running at :" + envs.port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Info("signal received: ", "sig", sig)

	tc, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.Shutdown(tc)
}

func createTemplates(envVars *Envs) map[string]*template.Template {
	templates := map[string]*template.Template{}
	templateFunctions := map[string]interface{}{
		"contains": func(xs []string, needle string) bool {
			for _, x := range xs {
				if x == needle {
					return true
				}
			}

			return false
		},
		"toTitle": func(s string) string {
			r := []rune(s)
			r[0] = unicode.ToUpper(r[0])

			return string(r)
		},
		"toLower": func(s string) string {
			return strings.ToLower(s)
		},
		"prefix": func(s string) string {
			return envVars.prefix + s
		},
		"sirius": func(s string) string {
			return envVars.siriusPublicURL + s
		},
		"showDirectDebits": func() bool {
			return envVars.showDirectDebits
		},
		"toCurrency": func(amount int) string {
			return shared.IntToDecimalString(amount)
		},
	}

	templateDirPath := filepath.Clean(envVars.webDir + "/template")
	templateDir, _ := os.Open(templateDirPath)
	templateDirs, _ := templateDir.Readdir(0)
	_ = templateDir.Close()

	mainTemplates, _ := filepath.Glob(templateDirPath + "/*.gotmpl")

	for _, file := range mainTemplates {
		tmpl := template.New(filepath.Base(file)).Funcs(templateFunctions)
		for _, dir := range templateDirs {
			if dir.IsDir() {
				tmpl, _ = tmpl.ParseGlob(templateDirPath + "/" + dir.Name() + "/*.gotmpl")
			}
		}
		tmpl, _ = tmpl.Parse(paginate.Template)
		templates[tmpl.Name()] = template.Must(tmpl.ParseFiles(file))
	}

	return templates
}
