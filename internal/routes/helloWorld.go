package routes

import (
	"fmt"
	"net/http"
)

func HandleHelloWorld() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello World!")
		},
	)
}
