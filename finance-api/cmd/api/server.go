package api

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Service interface {
	GetAccountInformation(ctx context.Context, id int32) (*shared.AccountInformation, error)
	GetInvoices(ctx context.Context, clientId int32) (*shared.Invoices, error)
	GetPermittedAdjustments(ctx context.Context, invoiceId int32) ([]shared.AdjustmentType, error)
	GetFeeReductions(ctx context.Context, invoiceId int32) (*shared.FeeReductions, error)
	AddInvoiceAdjustment(ctx context.Context, clientId int32, invoiceId int32, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error)
	GetInvoiceAdjustments(ctx context.Context, clientId int32) (*shared.InvoiceAdjustments, error)
	AddFeeReduction(ctx context.Context, clientId int32, data shared.AddFeeReduction) error
	CancelFeeReduction(ctx context.Context, id int32, cancelledFeeReduction shared.CancelFeeReduction) error
	UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int32, adjustmentId int32, status shared.AdjustmentStatus) error
	AddManualInvoice(ctx context.Context, clientId int32, invoice shared.AddManualInvoice) error
	GetBillingHistory(ctx context.Context, id int32) ([]shared.BillingHistory, error)
	ReapplyCredit(ctx context.Context, clientID int32) error
	UpdateClient(ctx context.Context, clientID int32, courtRef string) error
	ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error
	UpdatePaymentMethod(ctx context.Context, clientID int32, paymentMethod shared.PaymentMethod) error
}

type FileStorage interface {
	GetFileByVersion(ctx context.Context, bucketName string, filename string, versionID string) (*s3.GetObjectOutput, error)
	FileExists(ctx context.Context, bucketName string, filename string, versionID string) bool
}

type Reports interface {
	GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) error
}

type JWTClient interface {
	Verify(requestToken string) (*jwt.Token, error)
}

type Server struct {
	service     Service
	reports     Reports
	fileStorage FileStorage
	JWT         JWTClient
	validator   *validation.Validate
	envs        *Envs
}

type Envs struct {
	ReportsBucket     string
	GoLiveDate        time.Time
	EventBridgeAPIKey string
	SystemUserID      int32
}

func NewServer(service Service, reports Reports, fileStorage FileStorage, jwtClient JWTClient, validator *validation.Validate, envs *Envs) *Server {
	return &Server{
		service:     service,
		reports:     reports,
		fileStorage: fileStorage,
		JWT:         jwtClient,
		validator:   validator,
		envs:        envs,
	}
}

func (s *Server) SetupRoutes(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// authFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	authFunc := func(pattern string, h handlerFunc) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, s.authenticateAPI(handler))
	}
	authFunc("GET /clients/{clientId}", s.getAccountInformation)
	authFunc("GET /clients/{clientId}/invoices", s.getInvoices)
	authFunc("GET /clients/{clientId}/invoices/{invoiceId}/permitted-adjustments", s.getPermittedAdjustments)
	authFunc("GET /clients/{clientId}/fee-reductions", s.getFeeReductions)
	authFunc("GET /clients/{clientId}/invoice-adjustments", s.getInvoiceAdjustments)
	authFunc("GET /clients/{clientId}/billing-history", s.getBillingHistory)

	authFunc("POST /clients/{clientId}/invoices", s.addManualInvoice)
	authFunc("POST /clients/{clientId}/invoices/{invoiceId}/invoice-adjustments", s.AddInvoiceAdjustment)
	authFunc("PUT /clients/{clientId}/invoice-adjustments/{adjustmentId}", s.updatePendingInvoiceAdjustment)
	authFunc("POST /clients/{clientId}/fee-reductions", s.addFeeReduction)
	authFunc("PUT /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", s.cancelFeeReduction)
	authFunc("PUT /clients/{clientId}/payment-method", s.updatePaymentMethod)

	authFunc("GET /download", s.download)
	authFunc("HEAD /download", s.checkDownload)

	authFunc("POST /reports", s.requestReport)

	// unauthenticated as request is coming from EventBridge
	eventFunc := func(pattern string, h handlerFunc) {
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, s.authenticateEvent(handler))
	}
	eventFunc("POST /events", s.handleEvents)

	mux.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return otelhttp.NewHandler(telemetry.Middleware(logger)(securityheaders.Use(s.requestLogger(mux))), "supervision-finance-api")
}

func (s *Server) requestLogger(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health-check" {
			telemetry.LoggerFromContext(r.Context()).Info(
				"API Request",
				"method", r.Method,
				"uri", r.URL.RequestURI(),
			)
		}
		h.ServeHTTP(w, r)
	}
}

func (s *Server) getPathID(r *http.Request, key string) (int32, error) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 32)
	if err != nil {
		return 0, apierror.BadRequestError(key, "Unable to parse value to int", err)
	}
	if id < 1 {
		return 0, apierror.BadRequestError(key, "Invalid ID", nil)
	}
	return int32(id), nil
}
