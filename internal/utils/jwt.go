package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("jwt")
)

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "validating user is authenticated")
		defer span.End()

		token, err := r.Cookie("token")

		fmt.Println(token)

		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				HandleError(w, err, http.StatusBadRequest, "Token cookie provided", ctx)
			default:
				UnexpectedError(w, err, ctx)
			}
			return
		}

		ctx = context.WithValue(reqCtx, "token", token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
