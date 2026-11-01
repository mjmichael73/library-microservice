package server

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	// defaultLimit is used when the `limit` query parameter is missing or
	// invalid.
	defaultLimit = 10
	// maxLimit bounds the number of items a single list response may return so
	// a caller cannot force an unbounded result set.
	maxLimit = 100
)

// paginationParams reads the `limit` (default 10) and `offset` (default 0)
// query parameters. A missing, non-numeric, or non-positive `limit` falls back
// to defaultLimit and is capped at maxLimit; a missing, non-numeric, or
// negative `offset` falls back to 0.
func paginationParams(ctx echo.Context) (limit, offset int) {
	limit = defaultLimit
	offset = 0

	if v := ctx.QueryParam("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	if v := ctx.QueryParam("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

// paginate returns the sub-slice of items between offset and offset+limit. An
// offset at or past the end returns an empty, non-nil slice so the JSON
// envelope always carries an array rather than null.
func paginate[T any](items []T, offset, limit int) []T {
	if offset >= len(items) {
		return []T{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
