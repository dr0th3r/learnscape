package timetable

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerSubstitute = otel.Tracer("substitute timetable")
)

const substituteTimetableType = "substitute"

type SubstituteTimetable struct {
	id        int
	periodId  int
	subjectId int
	roomId    int
	schoolId  int
	date      time.Time
}

func ParseSubstituteTimetable(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing substitute timetable")

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
	dateUnprocessed := f.Get("date")
	date, err := time.Parse(time.DateOnly, dateUnprocessed)
	span.SetAttributes(
		attribute.String("date_unprocessed", dateUnprocessed),
		attribute.String("date", date.String()),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid date")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "substitute timetable", SubstituteTimetable{
		id:        -1,
		periodId:  periodId,
		subjectId: subjectId,
		roomId:    roomId,
		schoolId:  schoolId,
		date:      date,
	})

	return nil
}

func (t SubstituteTimetable) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(
		context.TODO(),
		`
		with inserted_timetable AS (
			insert into timetable (period_id, subject_id, room_id, school_id, type) 
			values ($1, $2, $3, $4, $5)
			returning id
		)
		insert into substitute_timetable (id, date)
		SELECT id, $6
		from inserted_timetable
		`,
		t.periodId, t.subjectId, t.roomId, t.schoolId, substituteTimetableType, t.date)
	return
}

func HandleCreateSubstituteTimetable(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracerSubstitute.Start(reqCtx, "create substitutet timetable")
			defer span.End()

			timetable := reqCtx.Value("substitute timetable").(SubstituteTimetable)

			if err := utils.HandleTx(ctx, db, timetable.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)

		},
	)
}
