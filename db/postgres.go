package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type pgxDB struct {
	conn *pgx.Conn
}

func newPgxDB(ctx context.Context, dsn string) (*pgxDB, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &pgxDB{conn: conn}, nil
}

func (d *pgxDB) ListTables(ctx context.Context) ([]Table, error) {
	query := `SELECT table_schema, table_name FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog', 'information_schema') ORDER BY table_schema, table_name`
	rows, err := d.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	table, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Table])
	if err != nil {
		return nil, err
	}

	return table, nil
}

func (d *pgxDB) ListColumns(ctx context.Context) ([]Column, error) {
	query := `SELECT table_schema, table_name, column_name
	          FROM information_schema.columns
	          WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
	          ORDER BY table_schema, table_name, ordinal_position`
	rows, err := d.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	cols, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Column])
	if err != nil {
		return nil, err
	}
	return cols, nil
}

func (d *pgxDB) FetchTableData(ctx context.Context, schema, table string, limit, offset int, sort *SortOption) (*QueryResult, error) {
	query := fmt.Sprintf("SELECT * FROM %s.%s", quotePostgresIdent(schema), quotePostgresIdent(table))
	if sort != nil && sort.Column != "" {
		direction := "ASC"
		if sort.Desc {
			direction = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", quotePostgresIdent(sort.Column), direction)
	}
	query += " LIMIT $1 OFFSET $2"
	rows, err := d.conn.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	fd := rows.FieldDescriptions()
	columns := make([]string, len(fd))
	for i, col := range fd {
		columns[i] = string(col.Name)
	}

	var resultRows [][]string
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make([]string, len(values))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		resultRows = append(resultRows, row)
	}

	return &QueryResult{Columns: columns, Rows: resultRows}, nil
}

func (d *pgxDB) ExecQuery(ctx context.Context, query string) (*QueryResult, error) {
	rows, err := d.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fd := rows.FieldDescriptions()
	columns := make([]string, len(fd))
	for i, col := range fd {
		columns[i] = string(col.Name)
	}

	var resultRows [][]string
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make([]string, len(values))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		resultRows = append(resultRows, row)
	}

	return &QueryResult{Columns: columns, Rows: resultRows}, nil
}

func (d *pgxDB) Close(ctx context.Context) error {
	return d.conn.Close(ctx)
}
