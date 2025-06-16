package db

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
)

type db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
}

type Client struct {
	db db
}

func NewClient(db *pgxpool.Pool) *Client {
	return &Client{db}
}

func (c *Client) Close() {
	c.db.Close()
}

type ReportQuery interface {
	GetQuery() string
	GetHeaders() []string
	GetParams() []any
	GetCallback() func(row pgx.CollectableRow) ([]string, error)
}

func NewReportQuery(query string) ReportQuery {
	return &reportQuery{query}
}

type reportQuery struct {
	Query string
}

func (q *reportQuery) GetQuery() string     { return q.Query }
func (q *reportQuery) GetHeaders() []string { return []string{} }
func (q *reportQuery) GetParams() []any     { return []any{} }
func (q *reportQuery) GetCallback() func(row pgx.CollectableRow) ([]string, error) {
	return func(row pgx.CollectableRow) ([]string, error) { return rowToStringMap(row) }
}

func (c *Client) Run(ctx context.Context, query ReportQuery) ([][]string, error) {
	headers := [][]string{query.GetHeaders()}

	rows, err := c.db.Query(ctx, query.GetQuery(), query.GetParams()...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	stringRows, err := pgx.CollectRows[[]string](rows, query.GetCallback())
	if err != nil {
		return nil, err
	}

	return append(headers, stringRows...), nil
}

func (c *Client) CopyStream(ctx context.Context, query ReportQuery) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		rows, err := c.db.Query(ctx, query.GetQuery(), query.GetParams()...)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		defer rows.Close()

		writer := csv.NewWriter(pw)
		defer writer.Flush()

		if err = writer.Write(query.GetHeaders()); err != nil {
			pw.CloseWithError(writer.Error())
			return
		}

		_, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) ([]string, error) {
			stringRow, err := query.GetCallback()(row)
			if err != nil {
				return nil, err
			}

			if err := writer.Write(stringRow); err != nil {
				return nil, err
			}
			return stringRow, nil
		})
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

func rowToStringMap(row pgx.CollectableRow) ([]string, error) {
	var stringRow []string
	values, err := row.Values()
	if err != nil {
		return nil, err
	}

	for _, value := range values {
		stringRow = append(stringRow, fmt.Sprintf("%v", value))
	}
	return stringRow, nil
}
