package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/dr0th3r/learnscape/internal/models"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterSchool(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "register school")
			defer span.End()

			school := reqCtx.Value("school").(models.School)
			admin := reqCtx.Value("user").(models.User)

			if err := utils.HandleTx(ctx, db, school.SaveToDB, admin.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}

			span.AddEvent("Starting to set jwt token for admin")
			if err := admin.SetToken(w, []byte("my secret"), time.Now().Add(time.Hour*72)); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error setting jwt", ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
func GetRegister() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, err := template.ParseFiles("./web/register.html"); err != nil {
				fmt.Println(w, err)
				return
			}

			tmpl := template.Must(template.ParseFiles("./web/register.html"))
			tmpl.Execute(w, nil)
		},
	)
}
