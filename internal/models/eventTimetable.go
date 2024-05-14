package models

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const eventTimetableType = "event"

type EventTimetable struct {
	id          int
	schoolId    int
	start       string
	end         string
	name        string
	description string
}

func ParseEventTimetable(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing event timetable")

	schoolId, err := utils.ParseInt(span, "school_id", f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid school id (not convertable to int)")
	}

	start := f.Get("start")
	_, err = time.Parse(time.RFC3339, start)
	span.SetAttributes(
		attribute.String("start_unprocessed", start),
		attribute.String("start", start),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid start time")
	}

	end := f.Get("end")
	_, err = time.Parse(time.RFC3339, end)
	span.SetAttributes(
		attribute.String("end_unprocessed", end),
		attribute.String("end", end),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid end time")
	}

	//TODO: add parsing time with setting attributes to utils
	//TODO: change start and end to time.Time

	name := f.Get("name")
	span.SetAttributes(
		attribute.String("name", name),
	)
	if name == "" {
		return utils.NewParserError(nil, "Name not provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "event timetable", EventTimetable{
		id:          -1,
		schoolId:    schoolId,
		start:       start,
		end:         end,
		name:        name,
		description: f.Get("description"),
	})

	return nil
}

func (t EventTimetable) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(
		context.TODO(),
		`
		WITH inserted_timetable AS (
		    INSERT INTO timetable (school_id, type) 
		    VALUES ($1, $2)
		    RETURNING id
		)
		INSERT INTO event_timetable (id, name, description, span)
		SELECT id, $3, $4, $5
		FROM inserted_timetable
		`,
		t.schoolId, eventTimetableType, t.name, t.description, fmt.Sprintf("[%s, %s]", t.start, t.end),
	)
	return
}
