package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// ══════════════════════════════════════════════════════════════════════════════
// CUSTOM STRUCT SCANNER (Zero Dependency)
// ══════════════════════════════════════════════════════════════════════════════
// Lightweight alternative to sqlx.Get/Select
// Uses stdlib database/sql only

// ScanRow scans a single row into a struct
// Supports db:"column_name" struct tags
//
// Usage:
//   row := db.QueryRowContext(ctx, query, args...)
//   var user User
//   err := ScanRow(row, &user)
func ScanRow(row *sql.Row, dest interface{}) error {
	// Get struct fields and their db tags
	fields, err := getStructFields(dest)
	if err != nil {
		return err
	}

	// Create slice of pointers for Scan
	scanDest := make([]interface{}, len(fields))
	for i := range fields {
		scanDest[i] = fields[i].Addr().Interface()
	}

	// Scan row into struct fields
	if err := row.Scan(scanDest...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	return nil
}

// ScanRows scans multiple rows into a slice of structs
//
// Usage:
//   rows, err := db.QueryContext(ctx, query, args...)
//   defer rows.Close()
//   var users []User
//   err := ScanRows(rows, &users)
func ScanRows(rows *sql.Rows, dest interface{}) error {
	// dest must be pointer to slice
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	sliceValue := destValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	// Get struct type from slice element type
	elemType := sliceValue.Type().Elem()

	// Get column names from rows
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Scan each row
	for rows.Next() {
		// Create new struct instance
		elemPtr := reflect.New(elemType)
		elem := elemPtr.Elem()

		// Get struct fields mapped to columns
		scanDest, err := getFieldPointers(elem, columns)
		if err != nil {
			return err
		}

		// Scan row into struct fields
		if err := rows.Scan(scanDest...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Append to slice
		sliceValue.Set(reflect.Append(sliceValue, elem))
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
// INTERNAL HELPERS
// ══════════════════════════════════════════════════════════════════════════════

// getStructFields returns addressable fields from a struct pointer
func getStructFields(dest interface{}) ([]reflect.Value, error) {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("dest must be a pointer to struct")
	}

	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("dest must be a pointer to struct")
	}

	structType := structValue.Type()
	numFields := structType.NumField()

	fields := make([]reflect.Value, 0, numFields)

	for i := 0; i < numFields; i++ {
		field := structType.Field(i)

		// Skip fields without db tag
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		fieldValue := structValue.Field(i)
		if !fieldValue.CanAddr() {
			continue
		}

		fields = append(fields, fieldValue)
	}

	return fields, nil
}

// getFieldPointers returns pointers to struct fields matching column names
func getFieldPointers(structValue reflect.Value, columns []string) ([]interface{}, error) {
	structType := structValue.Type()

	// Build map of column name -> field index
	fieldMap := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Handle tag format: db:"name,omitempty"
		tagName := strings.Split(dbTag, ",")[0]
		fieldMap[tagName] = i
	}

	// Create slice of field pointers in column order
	scanDest := make([]interface{}, len(columns))

	for i, column := range columns {
		fieldIndex, ok := fieldMap[column]
		if !ok {
			// Column not mapped to struct field - use dummy destination
			var dummy interface{}
			scanDest[i] = &dummy
			continue
		}

		field := structValue.Field(fieldIndex)
		if !field.CanAddr() {
			return nil, fmt.Errorf("field %s is not addressable", structType.Field(fieldIndex).Name)
		}

		scanDest[i] = field.Addr().Interface()
	}

	return scanDest, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// NULLABLE TYPES HELPER
// ══════════════════════════════════════════════════════════════════════════════

// NullString is a helper for nullable string fields
type NullString struct {
	sql.NullString
}

// Value returns the string value or empty string if null
func (n NullString) Value() string {
	if n.Valid {
		return n.String
	}
	return ""
}

// Ptr returns a pointer to the string value or nil if null
func (n NullString) Ptr() *string {
	if n.Valid {
		return &n.String
	}
	return nil
}

// NullInt64 is a helper for nullable int64 fields
type NullInt64 struct {
	sql.NullInt64
}

// Value returns the int64 value or 0 if null
func (n NullInt64) Value() int64 {
	if n.Valid {
		return n.Int64
	}
	return 0
}

// Ptr returns a pointer to the int64 value or nil if null
func (n NullInt64) Ptr() *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

// NullTime is a helper for nullable time fields
type NullTime struct {
	sql.NullTime
}

// Ptr returns a pointer to the time value or nil if null
func (n NullTime) Ptr() *interface{} {
	if n.Valid {
		var v interface{} = n.Time
		return &v
	}
	return nil
}
