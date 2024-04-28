package internal

import (
	r "github.com/dr0th3r/learnscape/internal/routes"
	"net/http"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	r.AddRoutes(mux)
	var handler http.Handler = mux
	return handler
}
