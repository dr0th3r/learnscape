package internal

import (
	hcheck "github.com/dr0th3r/learnscape/internal/healthCheck"
	"net/http"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux)
	var handler http.Handler = mux
	return handler
}

func addRoutes(
	mux *http.ServeMux,
) {
	mux.Handle("/health_check", hcheck.HandleHealthCheck())
}
