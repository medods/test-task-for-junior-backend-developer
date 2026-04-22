package task

import (
    "errors"
    "strings"
    "github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalidInput = errors.New("invalid task input")


func IsDuplicateError(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505"
    }
    return strings.Contains(err.Error(), "duplicate key value")
}
