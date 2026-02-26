package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type mysqlDB struct {
	conn *sql.DB
}

func newMysqlDB(dsn string) (*mysqlDB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return &mysqlDB{conn: conn}, nil
}

func (d *mysqlDB) ListTables(ctx context.Context) ([]Table, error) {
	query := `SELECT table_schema, table_name FROM information_schema.tables WHERE (DATABASE() IS NOT NULL AND table_schema = DATABASE()) OR (DATABASE() IS NULL AND table_schema NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')) ORDER BY table_schema, table_name`
	rows, err := d.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.Schema, &t.Name); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (d *mysqlDB) ListColumns(ctx context.Context) ([]Column, error) {
	query := `SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME
	          FROM INFORMATION_SCHEMA.COLUMNS
	          WHERE (DATABASE() IS NOT NULL AND TABLE_SCHEMA = DATABASE())
	             OR (DATABASE() IS NULL AND TABLE_SCHEMA NOT IN ('information_schema','mysql','performance_schema','sys'))
	          ORDER BY TABLE_NAME, ORDINAL_POSITION`
	rows, err := d.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []Column
	for rows.Next() {
		var c Column
		if err := rows.Scan(&c.Schema, &c.Table, &c.Name); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (d *mysqlDB) FetchTableData(ctx context.Context, schema, table string, limit, offset int, sort *SortOption) (*QueryResult, error) {
	query := fmt.Sprintf("SELECT * FROM %s.%s", quoteMySQLIdent(schema), quoteMySQLIdent(table))
	if sort != nil && sort.Column != "" {
		direction := "ASC"
		if sort.Desc {
			direction = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", quoteMySQLIdent(sort.Column), direction)
	}
	query += " LIMIT ? OFFSET ?"
	rows, err := d.conn.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var resultRows [][]string
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make([]string, len(columns))
		for i, v := range values {
			if v.Valid {
				row[i] = v.String
			} else {
				row[i] = "NULL"
			}
		}
		resultRows = append(resultRows, row)
	}

	return &QueryResult{Columns: columns, Rows: resultRows}, rows.Err()
}

func (d *mysqlDB) ExecQuery(ctx context.Context, query string) (*QueryResult, error) {
	rows, err := d.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var resultRows [][]string
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make([]string, len(columns))
		for i, v := range values {
			if v.Valid {
				row[i] = v.String
			} else {
				row[i] = "NULL"
			}
		}
		resultRows = append(resultRows, row)
	}

	return &QueryResult{Columns: columns, Rows: resultRows}, rows.Err()
}

func (d *mysqlDB) Close(_ context.Context) error {
	return d.conn.Close()
}
