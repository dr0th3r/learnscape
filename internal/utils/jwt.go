package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("jwt")
)

type UserClaims struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	SchoolId string `json:"schoolId"`
	jwt.RegisteredClaims
}

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "validating user is authenticated")
		defer span.End()

		tokenStr, err := r.Cookie("token")

		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				HandleError(w, err, http.StatusBadRequest, "Token cookie provided", ctx)
			default:
				UnexpectedError(w, err, ctx)
			}
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr.Value, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte("my secret"), nil
		})
		if err != nil {
			UnexpectedError(w, err, ctx)
		} else if claims, ok := token.Claims.(*UserClaims); !ok {
			UnexpectedError(w, err, ctx)
		} else {
			ctx = context.WithValue(reqCtx, "claims", claims)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
