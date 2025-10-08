package service

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Dispatch interface {
	CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error
	PaymentMethodChanged(ctx context.Context, event event.PaymentMethod) error
	DirectDebitScheduleFailed(ctx context.Context, event event.DirectDebitScheduleFailed) error
	RefundAdded(ctx context.Context, event event.RefundAdded) error
}

type FileStorage interface {
	GetFile(ctx context.Context, bucketName string, fileName string) (io.ReadCloser, error)
	GetFileWithVersion(ctx context.Context, bucketName string, fileName string, versionID string) (io.ReadCloser, error)
	StreamFile(ctx context.Context, bucketName string, fileName string, stream io.ReadCloser) (*string, error)
}

type NotifyClient interface {
	Send(ctx context.Context, payload notify.Payload) error
}

type AllpayClient interface {
	CancelMandate(ctx context.Context, data *allpay.CancelMandateRequest) error
	CreateMandate(ctx context.Context, data *allpay.CreateMandateRequest) error
	ModulusCheck(ctx context.Context, sortCode string, accountNumber string) error
	CreateSchedule(ctx context.Context, data *allpay.CreateScheduleInput) error
	FetchFailedPayments(ctx context.Context, input allpay.FetchFailedPaymentsInput) (allpay.FailedPayments, error)
}

type GovUKClient interface {
	AddWorkingDays(ctx context.Context, d time.Time, n int) (time.Time, error)
	SubWorkingDays(ctx context.Context, d time.Time, n int) (time.Time, error)
	NextWorkingDayOnOrAfterX(ctx context.Context, date time.Time, dayOfMonth int) (time.Time, error)
}

type Env struct {
	AsyncBucket string
}

type Service struct {
	store       *store.Queries
	dispatch    Dispatch
	fileStorage FileStorage
	notify      NotifyClient
	allpay      AllpayClient
	govUK       GovUKClient
	tx          TX
	env         *Env
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewService(conn *pgxpool.Pool, dispatch Dispatch, fileStorage FileStorage, notify NotifyClient, allpay AllpayClient, govUK GovUKClient, env *Env) *Service {
	return &Service{
		store:       store.New(conn),
		dispatch:    dispatch,
		fileStorage: fileStorage,
		notify:      notify,
		allpay:      allpay,
		govUK:       govUK,
		tx:          conn,
		env:         env,
	}
}

func (s *Service) BeginStoreTx(ctx context.Context) (*store.Tx, error) {
	tx, err := s.tx.Begin(ctx)
	logger := s.Logger(ctx)
	if err != nil {
		logger.Error("Could not begin a transaction", slog.String("err", err.Error()))
		return nil, err
	}

	return store.NewTx(tx), nil
}

func (s *Service) Logger(ctx context.Context) *slog.Logger {
	return telemetry.LoggerFromContext(ctx)
}
