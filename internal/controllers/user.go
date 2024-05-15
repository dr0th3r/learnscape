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

func RegisterUser(db *pgxpool.Pool) http.Handler {
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
			if err := user.SetToken(w, []byte("my secret"), time.Now().Add(time.Hour*72)); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error setting jwt", ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}

func Login(db *pgxpool.Pool) http.Handler {
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
			if err := user.SetToken(w, []byte("my secret"), time.Now().Add(time.Hour*72)); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error setting jwt", ctx)
			}

		},
	)
}

func GetLogin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./web/login.html"))
		tmpl.Execute(w, nil)
	})
}
