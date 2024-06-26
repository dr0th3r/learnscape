package controllers

import (
	"errors"
	"net/http"

	"github.com/dr0th3r/learnscape/internal/models"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePeriod(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "period creation")
			defer span.End()

			period := reqCtx.Value("period").(models.Period)

			claims := reqCtx.Value("claims").(*utils.UserClaims)

			if err := utils.HandleTx(ctx, db, period.SaveToDBWithSchoolId(claims.SchoolId)); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "22000" {
					utils.HandleError(w, err, http.StatusBadRequest, "Period times overlap or start is before end", ctx)
				} else {
					utils.UnexpectedError(w, err, ctx)
				}

				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
