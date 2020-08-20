package database

import (
	"context"
	"database/sql"
	"strconv"
)

type (
	ID      int64
	Decimal string
)

func (id ID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func ParseID(s string) (ID, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	return ID(v), err
}

type ContextRowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type ContextQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type ContextExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type ContextRowQueryExecutor interface {
	ContextRowQuerier
	ContextExecutor
}

type ContextQueryExecutor interface {
	ContextRowQuerier
	ContextQuerier
	ContextExecutor
}

type Scanner interface {
	Scan(dest ...interface{}) error
}
