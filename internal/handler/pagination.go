package handler

import (
	"net/http"
	"strconv"
)

func ParsePagination(r *http.Request) (limit, offset int) {
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}
