package room

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

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
	id        int
	name      string
	schoolId  int
	teacherId uuid.UUID
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing room")

	schoolIdUnprocessed := f.Get("school_id")
	span.SetAttributes(attribute.String("school_id_unprocessed", schoolIdUnprocessed))
	schoolId, err := strconv.Atoi(f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid school id")
	}
	span.SetAttributes(attribute.Int("school_id", schoolId))

	teacherIdUnprocessed := f.Get("teacher_id")
	span.SetAttributes(attribute.String("teacher_id_unprocessed", teacherIdUnprocessed))
	teacherId, err := uuid.Parse(teacherIdUnprocessed)
	if err != nil {
		return utils.NewParserError(err, "Invalid teacher id")
	}
	span.SetAttributes(attribute.String("teacher_id", teacherId.String()))

	name := f.Get("name")
	span.SetAttributes(attribute.String("name", name))
	if name == "" {
		return utils.NewParserError(nil, "Name not provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "room", Room{
		id:        -1,
		name:      name,
		schoolId:  schoolId,
		teacherId: teacherId,
	})

	return nil
}

func (r Room) SaveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "insert into room (name, school_id, teacher_id) values ($1, $2, $3)", r.name, r.schoolId, r.teacherId)
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
