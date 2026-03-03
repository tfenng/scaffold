package service

import (
	"testing"

	"github.com/jackc/pgconn"
	"github.com/tfenng/scaffold/internal/domain"
)

func TestMapUniqueViolation(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		wantCode   domain.Code
		wantMsg    string
	}{
		{
			name:       "uid unique",
			constraint: constraintUsersUIDUnique,
			wantCode:   domain.CodeConflict,
			wantMsg:    "uid already exists",
		},
		{
			name:       "name unique",
			constraint: constraintUsersNameUnique,
			wantCode:   domain.CodeConflict,
			wantMsg:    "name already exists",
		},
		{
			name:       "email unique partial index",
			constraint: constraintUsersEmailUnique,
			wantCode:   domain.CodeConflict,
			wantMsg:    "email already exists",
		},
		{
			name:       "email unique legacy constraint",
			constraint: constraintUsersEmailUniqueOld,
			wantCode:   domain.CodeConflict,
			wantMsg:    "email already exists",
		},
		{
			name:       "unknown constraint",
			constraint: "users_unknown_unique",
			wantCode:   domain.CodeConflict,
			wantMsg:    "unique constraint violation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := mapUniqueViolation(&pgconn.PgError{ConstraintName: tc.constraint})
			if err.Code != tc.wantCode {
				t.Fatalf("unexpected code: got=%s want=%s", err.Code, tc.wantCode)
			}
			if err.Message != tc.wantMsg {
				t.Fatalf("unexpected message: got=%q want=%q", err.Message, tc.wantMsg)
			}
		})
	}
}
