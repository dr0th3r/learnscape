package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func HandleError(w http.ResponseWriter, err error, code int, msg string, ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, msg)
	if err == nil && msg == "" {
		span.RecordError(errors.New("Unknown error, neither error nor msg were provided"))
		fmt.Fprintf(w, "An unexpected error occurred, please try again later")
		w.WriteHeader(code)
		return
	}

	if err == nil {
		span.RecordError(errors.New(msg))
	} else {
		span.RecordError(err)
	}

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

type ParseError struct {
	msg   string
	cause error
}

func NewParserError(cause error, msg string) *ParseError {
	return &ParseError{
		cause: cause,
		msg:   msg,
	}
}

func (e *ParseError) HandleError(w http.ResponseWriter, ctx context.Context) {
	HandleError(w, e.cause, http.StatusBadRequest, e.msg, ctx)
}
