package utils

import (
	"encoding/base64"
	"encoding/json"
	"go-web-template/internal/apperr"
)

const (
	defaultPageSize = 25
	maxPageSize     = 100
)

var errInvalidCursor = apperr.New(apperr.KindInvalid, "invalid cursor")

// Page is one page of a list. NextCursor is nil on the last page.
type Page[T any] struct {
	Data       []T     `json:"data"`
	NextCursor *string `json:"next_cursor"`
}

// PageSize returns the page size to use, with the default for a size out of range.
func PageSize(size int) int32 {
	if size < 1 || size > maxPageSize {
		return defaultPageSize
	}
	return int32(size) //nolint:gosec // size is in range above
}

// NewPage cuts rows to size. Fetch size+1 rows, so the extra row shows that a next page exists.
func NewPage[T any](rows []T, size int32, cursorOf func(T) any) (*Page[T], error) {
	if len(rows) <= int(size) {
		return &Page[T]{Data: rows}, nil
	}

	raw, err := json.Marshal(cursorOf(rows[size-1]))
	if err != nil {
		return nil, err
	}
	cursor := base64.RawURLEncoding.EncodeToString(raw)
	return &Page[T]{Data: rows[:size], NextCursor: &cursor}, nil
}

// DecodeCursor reads a cursor from NewPage into v.
func DecodeCursor(cursor string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil || json.Unmarshal(raw, v) != nil {
		return errInvalidCursor
	}
	return nil
}
