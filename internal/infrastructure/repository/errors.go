package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	errNotFound      = errors.New("resource not found")
	errAlreadyExists = errors.New("resource already exists")
	errInvalidInput  = errors.New("invalid input")
)

func handleDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return errAlreadyExists
		case "23503", "23502", "23514":
			return errInvalidInput
		}
	}
	return err
}
