package db

import "strings"

func quotePostgresIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

func quoteMySQLIdent(ident string) string {
	return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
}
