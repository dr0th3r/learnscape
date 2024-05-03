package utils

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func HandleError(w http.ResponseWriter, err error, code int, msg string, ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, msg)
	span.RecordError(err)

	w.WriteHeader(code)
	if msg == "" {
		fmt.Fprintf(w, err.Error())
	} else {
		fmt.Fprintf(w, msg)
	}
}

func UnexpectedError(w http.ResponseWriter, err error, ctx context.Context) {
	HandleError(w, err, http.StatusInternalServerError, "An unexpected error occurred, please try again later", ctx)
}
