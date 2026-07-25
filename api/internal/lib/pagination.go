package lib

import "github.com/gofiber/fiber/v3"

// Pagination defaults and bounds for list endpoints.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Page holds a validated limit/offset pair.
type Page struct {
	Limit  int
	Offset int
}

// ParsePage reads `limit` and `offset` query params, applying defaults and caps.
func ParsePage(c fiber.Ctx) Page {
	limit := fiber.Query(c, "limit", DefaultLimit)
	offset := fiber.Query(c, "offset", 0)

	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return Page{Limit: limit, Offset: offset}
}
