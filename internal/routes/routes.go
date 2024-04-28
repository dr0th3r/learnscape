package routes

import "net/http"

func AddRoutes(
	mux *http.ServeMux,
) {
	mux.Handle("/", handleHelloWorld())
}
