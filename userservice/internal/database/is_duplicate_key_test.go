package database

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"gorm duplicated key", gorm.ErrDuplicatedKey, true},
		{"pg error 23505", &pgconn.PgError{Code: "23505"}, true},
		{"pg error other code", &pgconn.PgError{Code: "23503"}, false},
		{"generic error", errors.New("something went wrong"), false},
		{"wrapped pg error", errors.Join(errors.New("wrap"), &pgconn.PgError{Code: "23505"}), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isDuplicateKeyError(tc.err)
			if got != tc.want {
				t.Fatalf("isDuplicateKeyError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsDuplicateKeyError_WrappedPgError(t *testing.T) {
	// errors.As must unwrap a wrapped *pgconn.PgError.
	err := &wrappedErr{inner: &pgconn.PgError{Code: "23505"}}
	if !isDuplicateKeyError(err) {
		t.Fatal("expected wrapped pg error to be detected")
	}
}

type wrappedErr struct {
	inner error
}

func (w *wrappedErr) Error() string { return "wrapped: " + w.inner.Error() }
func (w *wrappedErr) Unwrap() error { return w.inner }
