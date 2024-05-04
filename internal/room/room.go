package room

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("room")
)

type Room struct {
	id         int
	name       string
	school_id  uuid.UUID
	teacher_id uuid.UUID
}

func Parse(f url.Values, ctx context.Context) (Room, error) {
	span := trace.SpanFromContext(ctx)

	school_id, err := uuid.Parse(f.Get("school_id"))
	if err != nil {
		return Room{}, err
	}
	span.SetAttributes(attribute.String("school_id", school_id.String()))

	teacher_id, err := uuid.Parse(f.Get("teacher_id"))
	if err != nil {
		return Room{}, err
	}
	span.SetAttributes(attribute.String("teacher_id", teacher_id.String()))

	name := f.Get("name")
	if name == "" {
		return Room{}, errors.New("Missing field name")
	}

	return Room{
		id:         -1,
		name:       name,
		school_id:  school_id,
		teacher_id: teacher_id,
	}, nil
}

func (r Room) SaveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "insert into room (name, school_id, teacher_id) values ($1, $2, $3)", r.name, r.school_id, r.teacher_id)
	if err != nil {
		return err
	}
	return nil
}

func HandleCreateRoom(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "room creation")
			defer span.End()

			span.AddEvent("Starting to parse form data")
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing form data", ctx)
				return
			}

			span.AddEvent("Starting to parse room from form data")
			room, err := Parse(r.Form, ctx)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, "Invalid room", ctx)
				return
			}

			if err := utils.HandleTx(ctx, db, []utils.TxFunc{room.SaveToDB}); err != nil {
				utils.UnexpectedError(w, err, ctx)
				//TODO: add handling for invalid foreign keys later
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
