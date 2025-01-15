package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	GetHeaders() []string
	GetQuery() string
	GetParams() []any
}

func (c *Client) Run(ctx context.Context, query ReportQuery) ([][]string, error) {
	headers := [][]string{query.GetHeaders()}
	fmt.Println("Headers")
	fmt.Println(headers)

	rows, err := c.db.Query(ctx, query.GetQuery(), query.GetParams()...)
	fmt.Println("Rows raw")
	fmt.Println(rows)
	if err != nil {
		return nil, err
	}

	stringRows, err := pgx.CollectRows[[]string](rows, rowToStringMap)
	if err != nil {
		return nil, err
	}

	fmt.Println("Rows")
	fmt.Println(stringRows)

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
