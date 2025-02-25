package api

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log/slog"
	"net/http"
	"time"
)

type Service interface {
	GetAccountInformation(ctx context.Context, id int) (*shared.AccountInformation, error)
	GetInvoices(ctx context.Context, clientId int) (*shared.Invoices, error)
	GetPermittedAdjustments(ctx context.Context, invoiceId int) ([]shared.AdjustmentType, error)
	GetFeeReductions(ctx context.Context, invoiceId int) (*shared.FeeReductions, error)
	AddInvoiceAdjustment(ctx context.Context, clientId int, invoiceId int, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error)
	GetInvoiceAdjustments(ctx context.Context, clientId int) (*shared.InvoiceAdjustments, error)
	AddFeeReduction(ctx context.Context, clientId int, data shared.AddFeeReduction) error
	CancelFeeReduction(ctx context.Context, id int, cancelledFeeReduction shared.CancelFeeReduction) error
	UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int, adjustmentId int, status shared.AdjustmentStatus) error
	AddManualInvoice(ctx context.Context, clientId int, invoice shared.AddManualInvoice) error
	GetBillingHistory(ctx context.Context, id int) ([]shared.BillingHistory, error)
	ReapplyCredit(ctx context.Context, clientID int32) error
	UpdateClient(ctx context.Context, clientID int, courtRef string) error
	ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error
	UpdatePaymentMethod(ctx context.Context, clientID int, paymentMethod shared.PaymentMethod) error
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
	ReportsBucket string
	GoLiveDate    time.Time
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
	authFunc := func(pattern string, role string, h handlerFunc) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, s.requestLogger(s.authenticate(s.authorise(role)(handler))))
	}

	authFunc("GET /clients/{clientId}", shared.RoleAny, s.getAccountInformation)
	authFunc("GET /clients/{clientId}/invoices", shared.RoleAny, s.getInvoices)
	authFunc("GET /clients/{clientId}/invoices/{invoiceId}/permitted-adjustments", shared.RoleAny, s.getPermittedAdjustments)
	authFunc("GET /clients/{clientId}/fee-reductions", shared.RoleAny, s.getFeeReductions)
	authFunc("GET /clients/{clientId}/invoice-adjustments", shared.RoleAny, s.getInvoiceAdjustments)
	authFunc("GET /clients/{clientId}/billing-history", shared.RoleAny, s.getBillingHistory)

	authFunc("POST /clients/{clientId}/invoices", shared.RoleFinanceManager, s.addManualInvoice)
	authFunc("POST /clients/{clientId}/invoices/{invoiceId}/invoice-adjustments", shared.RoleFinanceUser, s.AddInvoiceAdjustment)
	authFunc("PUT /clients/{clientId}/invoice-adjustments/{adjustmentId}", shared.RoleFinanceManager, s.updatePendingInvoiceAdjustment)
	authFunc("POST /clients/{clientId}/fee-reductions", shared.RoleFinanceUser, s.addFeeReduction)
	authFunc("PUT /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", shared.RoleFinanceUser, s.cancelFeeReduction)
	authFunc("PUT /clients/{clientId}/payment-method", shared.RoleFinanceUser, s.updatePaymentMethod)

	authFunc("GET /download", shared.RoleFinanceReporting, s.download)
	authFunc("HEAD /download", shared.RoleFinanceReporting, s.checkDownload)

	authFunc("POST /reports", shared.RoleFinanceReporting, s.requestReport)

	// unauthenticated as request is coming from EventBridge
	eventFunc := func(pattern string, h handlerFunc) {
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, s.requestLogger(handler))
	}
	eventFunc("POST /events", s.handleEvents)

	mux.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return otelhttp.NewHandler(telemetry.Middleware(logger)(securityheaders.Use(mux)), "supervision-finance-api")
}

func (s *Server) requestLogger(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		telemetry.LoggerFromContext(r.Context()).Info(
			"API Request",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)
		h.ServeHTTP(w, r)
	}
}
