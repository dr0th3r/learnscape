package utils

import (
	"context"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel"
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
