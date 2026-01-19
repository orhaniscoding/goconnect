package database

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ══════════════════════════════════════════════════════════════════════════════
// TEST TYPES
// ══════════════════════════════════════════════════════════════════════════════

type TestUser struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	Age       int    `db:"age"`
	Active    bool   `db:"active"`
	CreatedAt time.Time `db:"created_at"`
}

type TestPartialUser struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	// Email field missing from struct
}

type TestUserWithOmitEmpty struct {
	ID       int64  `db:"id"`
	Name     string `db:"name,omitempty"`
	Email    string `db:"email"`
	Optional string `db:"optional,omitempty"`
}

// ══════════════════════════════════════════════════════════════════════════════
// ScanRow TESTS
// ══════════════════════════════════════════════════════════════════════════════

func TestScanRow_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "email", "age", "active", "created_at"}).
		AddRow(int64(1), "John Doe", "john@example.com", 30, true, now)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	row := db.QueryRow("SELECT * FROM users WHERE id = $1", 1)

	var user TestUser
	err = ScanRow(row, &user)
	if err != nil {
		t.Fatalf("ScanRow failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("expected Name 'John Doe', got '%s'", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("expected Email 'john@example.com', got '%s'", user.Email)
	}
	if user.Age != 30 {
		t.Errorf("expected Age 30, got %d", user.Age)
	}
	if !user.Active {
		t.Errorf("expected Active true, got %v", user.Active)
	}
}

func TestScanRow_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	row := db.QueryRow("SELECT * FROM users WHERE id = $1", 999)

	var user TestUser
	err = ScanRow(row, &user)
	if err == nil {
		t.Error("expected error for no rows, got nil")
	}
	// Scanner wraps sql.ErrNoRows, so we check for error existence
}

func TestScanRow_PartialStruct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Only return columns that match struct fields
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(int64(1), "Jane")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	row := db.QueryRow("SELECT * FROM users WHERE id = $1", 1)

	var user TestPartialUser
	err = ScanRow(row, &user)
	if err != nil {
		t.Fatalf("ScanRow failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Name != "Jane" {
		t.Errorf("expected Name 'Jane', got '%s'", user.Name)
	}
}

func TestScanRow_InvalidDest(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	tests := []struct {
		name string
		dest interface{}
	}{
		{"nil dest", nil},
		{"non-pointer", TestUser{}},
		{"pointer to non-struct", new(int)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := db.QueryRow("SELECT 1")
			err := ScanRow(row, tt.dest)
			if err == nil {
				t.Error("expected error for invalid dest, got nil")
			}
		})
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// ScanRows TESTS
// ══════════════════════════════════════════════════════════════════════════════

func TestScanRows_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "email", "age", "active", "created_at"}).
		AddRow(int64(1), "John Doe", "john@example.com", 30, true, now).
		AddRow(int64(2), "Jane Smith", "jane@example.com", 25, false, now)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rawRows, err := db.Query("SELECT * FROM users")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rawRows.Close()

	var users []TestUser
	err = ScanRows(rawRows, &users)
	if err != nil {
		t.Fatalf("ScanRows failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].ID != 1 {
		t.Errorf("expected first user ID 1, got %d", users[0].ID)
	}
	if users[1].ID != 2 {
		t.Errorf("expected second user ID 2, got %d", users[1].ID)
	}
}

func TestScanRows_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "age", "active", "created_at"})

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rawRows, err := db.Query("SELECT * FROM users WHERE active = $1", false)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rawRows.Close()

	var users []TestUser
	err = ScanRows(rawRows, &users)
	if err != nil {
		t.Fatalf("ScanRows failed: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestScanRows_InvalidDest(t *testing.T) {
	tests := []struct {
		name string
		dest interface{}
	}{
		{"non-pointer", []TestUser{}},
		{"pointer to non-slice", new(int)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh DB for each test
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			// Setup valid rows
			rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test")
			mock.ExpectQuery("SELECT").WillReturnRows(rows)

			rawRows, err := db.Query("SELECT 1")
			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}

			err = ScanRows(rawRows, tt.dest)
			if err == nil {
				t.Error("expected error for invalid dest, got nil")
			}
			rawRows.Close()
		})
	}
}

func TestScanRows_WithOmitEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "optional"}).
		AddRow(int64(1), "Test", "test@example.com", "value")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rawRows, err := db.Query("SELECT * FROM users")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rawRows.Close()

	var users []TestUserWithOmitEmpty
	err = ScanRows(rawRows, &users)
	if err != nil {
		t.Fatalf("ScanRows failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Optional != "value" {
		t.Errorf("expected Optional 'value', got '%s'", users[0].Optional)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// NULLABLE TYPES TESTS
// ══════════════════════════════════════════════════════════════════════════════

func TestNullString_Value(t *testing.T) {
	tests := []struct {
		name string
		ns   NullString
		want string
	}{
		{
			name: "valid string",
			ns:   NullString{sql.NullString{String: "test", Valid: true}},
			want: "test",
		},
		{
			name: "null string",
			ns:   NullString{sql.NullString{String: "", Valid: false}},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ns.Value(); got != tt.want {
				t.Errorf("NullString.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullString_Ptr(t *testing.T) {
	tests := []struct {
		name string
		ns   NullString
		want *string
	}{
		{
			name: "valid string",
			ns:   NullString{sql.NullString{String: "test", Valid: true}},
			want: func() *string { s := "test"; return &s }(),
		},
		{
			name: "null string",
			ns:   NullString{sql.NullString{String: "", Valid: false}},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ns.Ptr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NullString.Ptr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullInt64_Value(t *testing.T) {
	tests := []struct {
		name string
		ni   NullInt64
		want int64
	}{
		{
			name: "valid int64",
			ni:   NullInt64{sql.NullInt64{Int64: 42, Valid: true}},
			want: 42,
		},
		{
			name: "null int64",
			ni:   NullInt64{sql.NullInt64{Int64: 0, Valid: false}},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ni.Value(); got != tt.want {
				t.Errorf("NullInt64.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullInt64_Ptr(t *testing.T) {
	tests := []struct {
		name string
		ni   NullInt64
		want *int64
	}{
		{
			name: "valid int64",
			ni:   NullInt64{sql.NullInt64{Int64: 42, Valid: true}},
			want: func() *int64 { i := int64(42); return &i }(),
		},
		{
			name: "null int64",
			ni:   NullInt64{sql.NullInt64{Int64: 0, Valid: false}},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ni.Ptr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NullInt64.Ptr() = %v, want %v", got, tt.want)
			}
		})
	}
}
