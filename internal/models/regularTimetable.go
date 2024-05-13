package models

import (
	"context"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const regularTimetableType = "regular"

type RegularTimetable struct {
	id        int
	periodId  int
	subjectId int
	roomId    int
	schoolId  int
	weekday   string
}

func ParseRegularTimetable(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing regular timetable")

	weekday := f.Get("weekday")
	span.SetAttributes(
		attribute.String("weekday", weekday),
	)

	switch weekday {
	case "1":
		weekday = "Po"
	case "2":
		weekday = "Út"
	case "3":
		weekday = "St"
	case "4":
		weekday = "Čt"
	case "5":
		weekday = "Pá"
	default:
		return utils.NewParserError(nil, "Invalid weekday")
	}

	periodId, err := utils.ParseInt(span, "period_id", f.Get("period_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid period id (not convertable to int)")
	}
	subjectId, err := utils.ParseInt(span, "subject_id", f.Get("subject_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid subject id (not convertable to int)")
	}
	roomId, err := utils.ParseInt(span, "room_id", f.Get("room_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid room id (not convertable to int)")
	}
	schoolId, err := utils.ParseInt(span, "school_id", f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid school id (not convertable to int)")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "regular timetable", RegularTimetable{
		id:        -1,
		periodId:  periodId,
		subjectId: subjectId,
		roomId:    roomId,
		schoolId:  schoolId,
		weekday:   weekday,
	})

	return nil
}

func (t RegularTimetable) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(
		context.TODO(),
		`
		WITH inserted_timetable AS (
		    INSERT INTO timetable (school_id, type) 
		    VALUES ($1, $2)
		    RETURNING id
		),
		inserted_academic_timetable AS (
		    INSERT INTO academic_timetable (id, period_id, subject_id, room_id)
		    SELECT id, $3, $4, $5
		    FROM inserted_timetable
		)
		INSERT INTO regular_timetable (id, weekday)
		SELECT id, $6
		FROM inserted_timetable
		`,
		t.schoolId, regularTimetableType, t.periodId, t.subjectId, t.roomId, t.weekday)

	return
}
