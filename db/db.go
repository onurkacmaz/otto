package db

import "context"

type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverMySQL    Driver = "mysql"
)

type DB interface {
	ListTables(ctx context.Context) ([]Table, error)
	FetchTableData(ctx context.Context, schema, table string, limit, offset int) (*QueryResult, error)
	ExecQuery(ctx context.Context, query string) (*QueryResult, error)
	Close(ctx context.Context) error
}

type Table struct {
	Schema string
	Name   string
}

type QueryResult struct {
	Columns []string
	Rows    [][]string
}
