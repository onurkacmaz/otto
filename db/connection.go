package db

import (
	"context"
	"fmt"
)

type Config struct {
	Driver   Driver `json:"driver,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	DBName   string `json:"dbname,omitempty"`
}

func (c Config) DSN() string {
	host := c.Host
	if host == "" {
		host = "localhost"
	}
	user := c.User
	dbname := c.DBName

	if c.Driver == DriverMySQL {
		if user == "" {
			user = "root"
		}
		port := c.Port
		if port == "" {
			port = "3306"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, c.Password, host, port, dbname)
	}

	if user == "" {
		user = "postgres"
	}
	port := c.Port
	if port == "" {
		port = "5432"
	}
	if dbname == "" {
		dbname = "postgres"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, c.Password, host, port, dbname)
}

func Connect(ctx context.Context, cfg Config) (DB, error) {
	switch cfg.Driver {
	case DriverMySQL:
		return newMysqlDB(cfg.DSN())
	default:
		return newPgxDB(ctx, cfg.DSN())
	}
}
