package database

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mjmichael73/library-microservice/userservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/userservice/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockUserClient(t *testing.T) (Client, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return Client{DB: db}, mock
}

// TestUpdateUser_NotFound covers R-034: GORM Updates does not error on 0 rows,
// so the RowsAffected check must surface a NotFoundError.
func TestUpdateUser_NotFound(t *testing.T) {
	client, mock := newMockUserClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET "first_name"=\$1,"last_name"=\$2 WHERE "users"."user_id" = \$3`).
		WithArgs("Grace", "Hopper", "missing-id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	_, err := client.UpdateUser(context.Background(), &models.User{UserID: "missing-id", FirstName: "Grace", LastName: "Hopper"})
	if _, ok := err.(*dberrors.NotFoundError); !ok {
		t.Fatalf("expected *dberrors.NotFoundError, got %T (%v)", err, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestUpdateUser_Success covers R-034: the updated row is re-fetched by id.
func TestUpdateUser_Success(t *testing.T) {
	client, mock := newMockUserClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET "first_name"=\$1,"last_name"=\$2 WHERE "users"."user_id" = \$3`).
		WithArgs("Grace", "Hopper", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."user_id" = \$1 ORDER BY "users"."user_id" LIMIT \$2`).
		WithArgs("user-1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "first_name", "last_name", "email", "password", "is_admin"}).
			AddRow("user-1", "Grace", "Hopper", "grace@example.com", "hash", false))

	user, err := client.UpdateUser(context.Background(), &models.User{UserID: "user-1", FirstName: "Grace", LastName: "Hopper"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.FirstName != "Grace" || user.LastName != "Hopper" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
