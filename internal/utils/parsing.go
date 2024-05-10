package utils

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracerParser = otel.Tracer("parser")

type parserFunc func(url.Values, context.Context, *context.Context) *ParseError

type parser struct {
	argName    string
	parserFunc parserFunc
}

func NewParser(
	argName string,
	parserFunc parserFunc,
) parser {
	return parser{
		argName:    argName,
		parserFunc: parserFunc,
	}
}

func ParseForm(next http.Handler, parserFuncs ...parserFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		parserCtx, span := tracerParser.Start(reqCtx, "parsing formdata")
		defer span.End()

		span.AddEvent("Parsing form data")
		if err := r.ParseForm(); err != nil {
			HandleError(w, err, http.StatusInternalServerError, "Error parsing formdata", parserCtx)
			return
		}

		handlerCtx := reqCtx

		for _, f := range parserFuncs {
			//parser adds parsed values to handlerCtx
			if err := f(r.Form, parserCtx, &handlerCtx); err != nil {
				err.HandleError(w, parserCtx)
				return
			}
		}

		req := r.WithContext(handlerCtx)

		next.ServeHTTP(w, req)
	})
}

func ParseInt(span trace.Span, key, value string) (int, error) {
	intValue, err := strconv.Atoi(value)
	span.SetAttributes(
		attribute.String(key+"_unprocessed", value),
		attribute.Int(key, intValue),
	)

	if err != nil {
		return 0, err
	}

	return intValue, nil
}

func ParseUuid(span trace.Span, key, value string) (uuid.UUID, error) {
	uuid, err := uuid.Parse(value)
	span.SetAttributes(
		attribute.String(key+"_unprocessed", value),
		attribute.String(key, uuid.String()),
	)
	return uuid, err
}

func ParseTime(span trace.Span, key, value, format string) (time.Time, error) {
	t, err := time.Parse(format, value)
	span.SetAttributes(
		attribute.String(key+"_unprocessed", value),
		attribute.String(key, t.String()),
	)
	return t, err
}
