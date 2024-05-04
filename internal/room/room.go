package room

import (
	"context"
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

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing room")

	school_id, err := uuid.Parse(f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid school id")
	}
	span.SetAttributes(attribute.String("school_id", school_id.String()))

	teacher_id, err := uuid.Parse(f.Get("teacher_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid teacher id")
	}
	span.SetAttributes(attribute.String("teacher_id", teacher_id.String()))

	name := f.Get("name")
	if name == "" {
		return utils.NewParserError(nil, "Name not provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "room", Room{
		id:         -1,
		name:       name,
		school_id:  school_id,
		teacher_id: teacher_id,
	})

	return nil
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
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "room creation")
			defer span.End()

			room := reqCtx.Value("room").(Room)

			if err := utils.HandleTx(ctx, db, room.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				//TODO: add handling for invalid foreign keys later
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
