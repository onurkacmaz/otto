package db

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
)

const nullDisplayValue = "NULL"

func formatMySQLValue(value any, typeName string) string {
	switch v := value.(type) {
	case nil:
		return nullDisplayValue
	case []byte:
		return formatBytesForDisplay(v, typeName, false)
	default:
		return formatScalarValue(v)
	}
}

func formatPostgresValue(value any, oid uint32) string {
	switch v := value.(type) {
	case nil:
		return nullDisplayValue
	case [16]byte:
		if oid == pgtype.UUIDOID {
			return formatUUIDBytes(v[:])
		}
		return formatScalarValue(v)
	case []byte:
		return formatBytesForDisplay(v, "", oid == pgtype.UUIDOID)
	case pgtype.UUID:
		if !v.Valid {
			return nullDisplayValue
		}
		return formatUUIDBytes(v.Bytes[:])
	default:
		return formatScalarValue(v)
	}
}

func formatScalarValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func formatBytesForDisplay(raw []byte, typeName string, forceUUID bool) string {
	if len(raw) == 0 {
		return ""
	}
	if forceUUID || shouldFormatAsUUID(typeName, raw) {
		return formatUUIDBytes(raw)
	}
	if isTextualColumnType(typeName) || isReadableText(raw) {
		return string(raw)
	}
	return "0x" + hex.EncodeToString(raw)
}

func shouldFormatAsUUID(typeName string, raw []byte) bool {
	if len(raw) != 16 {
		return false
	}
	switch strings.ToUpper(typeName) {
	case "BINARY", "VARBINARY", "UUID":
		return true
	default:
		return false
	}
}

func isTextualColumnType(typeName string) bool {
	switch strings.ToUpper(typeName) {
	case "CHAR", "VARCHAR", "TEXT", "TINYTEXT", "MEDIUMTEXT", "LONGTEXT", "JSON", "ENUM", "SET":
		return true
	default:
		return false
	}
}

func isReadableText(raw []byte) bool {
	if !utf8.Valid(raw) {
		return false
	}
	for _, r := range string(raw) {
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	return true
}

func formatUUIDBytes(raw []byte) string {
	if len(raw) != 16 {
		return "0x" + hex.EncodeToString(raw)
	}

	const hexDigits = "0123456789abcdef"
	buf := make([]byte, 36)
	j := 0
	for i, b := range raw {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			buf[j] = '-'
			j++
		}
		buf[j] = hexDigits[b>>4]
		buf[j+1] = hexDigits[b&0x0f]
		j += 2
	}

	return string(buf)
}
