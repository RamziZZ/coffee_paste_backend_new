package handler

import (
	"io"
	"net/http"
	"strings"

	"coffeepaste/internal/middleware"
	"coffeepaste/internal/model"
)

func stringsReader(s string) io.Reader { return strings.NewReader(s) }

func withUser(r *http.Request) *http.Request {
	return r.WithContext(middleware.TestContextWithUser(r.Context(), &model.User{ID: 1, Name: "T", Role: "customer"}))
}
