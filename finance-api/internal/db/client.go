package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBClient interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
}

type Client struct {
	db DBClient
}

func NewClient(db *pgxpool.Pool) *Client {
	return &Client{db}
}

func (c *Client) Close() {
	c.db.Close()
}

type ReportQuery interface {
	GetHeaders() []string
	GetQuery() string
	GetParams() []any
}

func (c *Client) Run(ctx context.Context, query ReportQuery) ([][]string, error) {
	headers := [][]string{query.GetHeaders()}

	rows, err := c.db.Query(ctx, query.GetQuery(), query.GetParams()...)
	if err != nil {
		return nil, err
	}

	stringRows, err := pgx.CollectRows[[]string](rows, rowToStringMap)
	if err != nil {
		return nil, err
	}

	return append(headers, stringRows...), nil
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
