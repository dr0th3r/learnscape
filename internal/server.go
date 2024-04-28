package internal

import (
	r "github.com/dr0th3r/learnscape/internal/routes"
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
	mux.Handle("/", r.HandleHelloWorld())
}
