package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestHandleDBError_NoRows(t *testing.T) {
	err := handleDBError(pgx.ErrNoRows)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestHandleDBError_UniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code: "23505",
	}
	err := handleDBError(pgErr)
	assert.ErrorIs(t, err, ErrAlreadyExists)
}

func TestHandleDBError_ForeignKeyViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code: "23503",
	}
	err := handleDBError(pgErr)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestHandleDBError_NotNullViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code: "23502",
	}
	err := handleDBError(pgErr)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestHandleDBError_CheckViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code: "23514",
	}
	err := handleDBError(pgErr)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestHandleDBError_UnknownError(t *testing.T) {
	unknownErr := errors.New("unknown error")
	err := handleDBError(unknownErr)
	assert.Equal(t, unknownErr, err)
}

func TestHandleDBError_Nil(t *testing.T) {
	err := handleDBError(nil)
	assert.NoError(t, err)
}

