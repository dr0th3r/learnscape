package controllers

import (
	"net/http"

	"github.com/dr0th3r/learnscape/internal/models"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTimetableGroup(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "create timetable group")
		defer span.End()

		usersGroup := reqCtx.Value("timetable_group").(models.TimetableGroup)

		err := utils.HandleTx(ctx, db, usersGroup.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
