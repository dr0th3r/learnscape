package models

import (
	"context"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Room struct {
	id        int
	name      string
	schoolId  int
	teacherId uuid.UUID
}

func ParseRoom(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing room")

	schoolId, err := utils.ParseInt(span, "school_id", f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid school id")
	}
	teacherId, err := utils.ParseUuid(span, "teacher_id", f.Get("teacher_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid teacher id")
	}
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
