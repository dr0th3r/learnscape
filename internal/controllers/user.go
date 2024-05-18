package controllers

import (
	"errors"
	"html/template"
	"net/http"
	"time"

	"github.com/dr0th3r/learnscape/internal/models"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const jwtCookieLifetime = time.Hour * 72

func RegisterUser(db *pgxpool.Pool, jwtSecret string) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "user registration")
			defer span.End()

			user := reqCtx.Value("user").(models.User)
			if !user.HasSchoolId() {
				utils.NewParserError(nil, "User doesn't have school id").HandleError(w, ctx)
				return
			}

			if err := utils.HandleTx(ctx, db, user.SaveToDB); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					utils.HandleError(w, err, http.StatusConflict, "Email already registered", ctx)
				} else {
					utils.UnexpectedError(w, err, ctx)
				}
				return
			}

			span.AddEvent("Set user jwt token")
			tokenCookie, err := user.CreateTokenCookie([]byte(jwtSecret), time.Now().Add(jwtCookieLifetime))
			if err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}
			http.SetCookie(w, tokenCookie)

			w.WriteHeader(http.StatusCreated)
		},
	)
}

func Login(db *pgxpool.Pool, jwtSecret string) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "user login")
			defer span.End()

			user := reqCtx.Value("user").(models.User)

			span.AddEvent("Log user in")
			if err := user.Login(db); err != nil {
				utils.HandleError(w, err, http.StatusUnauthorized, "Failed to log user in", ctx)
				return
			}

			span.AddEvent("Set user jwt")
			tokenCookie, err := user.CreateTokenCookie([]byte(jwtSecret), time.Now().Add(jwtCookieLifetime))
			if err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}
			http.SetCookie(w, tokenCookie)

		},
	)
}

func GetLogin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./web/login.html"))
		tmpl.Execute(w, nil)
	})
}
