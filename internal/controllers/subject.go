package controllers

import (
	"net/http"

	"github.com/dr0th3r/learnscape/internal/models"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateSubject(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "subject creation")
			defer span.End()

			subject := reqCtx.Value("subject").(models.Subject)

			if err := utils.HandleTx(ctx, db, subject.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
