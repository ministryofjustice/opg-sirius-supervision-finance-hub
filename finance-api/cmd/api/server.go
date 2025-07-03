package api

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Service interface {
	GetAccountInformation(ctx context.Context, id int32) (*shared.AccountInformation, error)
	GetInvoices(ctx context.Context, clientId int32) (shared.Invoices, error)
	GetInvoiceAdjustments(ctx context.Context, clientId int32) (shared.InvoiceAdjustments, error)
	GetPermittedAdjustments(ctx context.Context, invoiceId int32) ([]shared.AdjustmentType, error)
	GetRefunds(ctx context.Context, clientId int32) (shared.Refunds, error)
	GetFeeReductions(ctx context.Context, invoiceId int32) (shared.FeeReductions, error)
	AddInvoiceAdjustment(ctx context.Context, clientId int32, invoiceId int32, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error)
	AddFeeReduction(ctx context.Context, clientId int32, data shared.AddFeeReduction) error
	CancelFeeReduction(ctx context.Context, id int32, cancelledFeeReduction shared.CancelFeeReduction) error
	UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int32, adjustmentId int32, status shared.AdjustmentStatus) error
	AddManualInvoice(ctx context.Context, clientId int32, invoice shared.AddManualInvoice) error
	GetBillingHistory(ctx context.Context, id int32) ([]shared.BillingHistory, error)
	ReapplyCredit(ctx context.Context, clientID int32) error
	UpdateClient(ctx context.Context, clientID int32, courtRef string) error
	UpdatePaymentMethod(ctx context.Context, clientID int32, paymentMethod shared.PaymentMethod) error
	ProcessPayments(ctx context.Context, records [][]string, uploadType shared.ReportUploadType, bankDate shared.Date, pisNumber int) (map[int]string, error)
	ProcessAdhocEvent(ctx context.Context) error
	ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType) (map[int]string, error)
	AddRefund(ctx context.Context, clientId int32, refund shared.AddRefund) error
	UpdateRefundDecision(ctx context.Context, clientId int32, refundId int32, status shared.RefundStatus) error
	PostReportActions(ctx context.Context, report shared.ReportRequest)
	ProcessFulfilledRefunds(ctx context.Context, records [][]string, date shared.Date) (map[int]string, error)
	ExpireRefunds(ctx context.Context) error
	ProcessDirectUploadReport(ctx context.Context, filename string, fileBytes io.Reader, uploadType shared.ReportUploadType) error
}
type FileStorage interface {
	GetFile(ctx context.Context, bucketName string, filename string) (io.ReadCloser, error)
	GetFileWithVersion(ctx context.Context, bucketName string, filename string, versionID string) (io.ReadCloser, error)
	FileExists(ctx context.Context, bucketName string, filename string) bool
	FileExistsWithVersion(ctx context.Context, bucketName string, filename string, versionID string) bool
}

type Reports interface {
	GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time)
}

type JWTClient interface {
	Verify(requestToken string) (*jwt.Token, error)
}

type NotifyClient interface {
	Send(ctx context.Context, payload notify.Payload) error
}

type Server struct {
	service           Service
	reports           Reports
	fileStorage       FileStorage
	notify            NotifyClient
	JWT               JWTClient
	validator         *validation.Validate
	envs              *Envs
	onReportRequested func() // hook to allow tests to wait on async function to complete
}

type Envs struct {
	ReportsBucket     string
	GoLiveDate        time.Time
	EventBridgeAPIKey string
	SystemUserID      int32
}

func NewServer(service Service, reports Reports, fileStorage FileStorage, notify NotifyClient, jwtClient JWTClient, validator *validation.Validate, envs *Envs) *Server {
	return &Server{
		service:     service,
		reports:     reports,
		fileStorage: fileStorage,
		notify:      notify,
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
		mux.Handle(pattern, s.authenticateAPI(s.requestLogger(s.authorise(role)(handler))))
	}

	authFunc("GET /clients/{clientId}", shared.RoleAny, s.getAccountInformation)
	authFunc("GET /clients/{clientId}/invoices", shared.RoleAny, s.getInvoices)
	authFunc("GET /clients/{clientId}/invoices/{invoiceId}/permitted-adjustments", shared.RoleAny, s.getPermittedAdjustments)
	authFunc("GET /clients/{clientId}/fee-reductions", shared.RoleAny, s.getFeeReductions)
	authFunc("GET /clients/{clientId}/invoice-adjustments", shared.RoleAny, s.getInvoiceAdjustments)
	authFunc("GET /clients/{clientId}/billing-history", shared.RoleAny, s.getBillingHistory)
	authFunc("GET /clients/{clientId}/refunds", shared.RoleAny, s.getRefunds)

	authFunc("POST /clients/{clientId}/invoices", shared.RoleFinanceManager, s.addManualInvoice)
	authFunc("POST /clients/{clientId}/invoices/{invoiceId}/invoice-adjustments", shared.RoleFinanceUser, s.AddInvoiceAdjustment)
	authFunc("PUT /clients/{clientId}/invoice-adjustments/{adjustmentId}", shared.RoleFinanceManager, s.updatePendingInvoiceAdjustment)
	authFunc("POST /clients/{clientId}/fee-reductions", shared.RoleFinanceUser, s.addFeeReduction)
	authFunc("PUT /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", shared.RoleFinanceManager, s.cancelFeeReduction)
	authFunc("PUT /clients/{clientId}/payment-method", shared.RoleFinanceUser, s.updatePaymentMethod)
	authFunc("POST /clients/{clientId}/refunds", shared.RoleFinanceUser, s.addRefund)
	authFunc("PUT /clients/{clientId}/refunds/{refundId}", shared.RoleFinanceManager, s.updateRefundDecision)

	authFunc("GET /download", shared.RoleFinanceReporting, s.download)
	authFunc("HEAD /download", shared.RoleFinanceReporting, s.checkDownload)

	authFunc("POST /reports", shared.RoleFinanceReporting, s.requestReport)

	authFunc("POST /uploads", shared.RoleFinanceReporting, s.processUpload)

	// unauthenticated as request is coming from EventBridge
	eventFunc := func(pattern string, h handlerFunc) {
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, s.authenticateEvent(s.requestLogger(handler)))
	}
	eventFunc("POST /events", s.handleEvents)

	mux.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return otelhttp.NewHandler(telemetry.Middleware(logger)(securityheaders.Use(mux)), "supervision-finance-api")
}

func (s *Server) requestLogger(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().(auth.Context)
		s.Logger(ctx).Info(
			"API Request",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"user-id", ctx.User.ID,
		)
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

func (s *Server) Logger(ctx context.Context) *slog.Logger {
	return telemetry.LoggerFromContext(ctx)
}

func (s *Server) copyCtx(r *http.Request) context.Context {
	copyCtx := telemetry.ContextWithLogger(context.Background(), s.Logger(r.Context()))
	return auth.Context{
		Context: copyCtx,
		User:    r.Context().(auth.Context).User,
	}
}

// unchecked allows errors to be unchecked when deferring a function, e.g. closing a reader where a failure would only
// occur when the process is likely to already be unrecoverable
func unchecked(f func() error) {
	_ = f()
}
