package controllers

import (
	"net/http"

	"github.com/dr0th3r/learnscape/internal/models"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRoom(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "room creation")
			defer span.End()

			room := reqCtx.Value("room").(models.Room)

			if err := utils.HandleTx(ctx, db, room.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				//TODO: add handling for invalid foreign keys later
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
