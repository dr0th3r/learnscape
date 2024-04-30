package utils

import (
	"fmt"
	"net/http"
)

func HandleError(w http.ResponseWriter, err error, code int, msg string) {
	fmt.Printf(err.Error())
	w.WriteHeader(code)
	if msg == "" {
		fmt.Fprintf(w, err.Error())
	} else {
		fmt.Fprintf(w, msg)
	}
}
