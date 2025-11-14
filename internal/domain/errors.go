package domain

import "errors"

var (
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrNotFound       = errors.New("not found")
	ErrNoRowsAffected = errors.New("no rows affected")
)
