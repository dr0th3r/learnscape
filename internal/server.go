package internal

import "net/http"

func NewServer() http.Handler {
	mux := http.NewServeMux()
	var handler http.Handler = mux
	return handler
}
