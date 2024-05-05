package timetable

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("regular timetable")
)

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

	periodIdUnprocessed := f.Get("period_id")
	span.SetAttributes(attribute.String("period_id_unprocessed", periodIdUnprocessed))
	periodId, err := strconv.Atoi(periodIdUnprocessed)
	if err != nil {
		return utils.NewParserError(nil, "Invalid period id (not convertable to int)")
	}
	span.SetAttributes(attribute.Int("period_id", periodId))

	subjectIdUnprocessed := f.Get("subject_id")
	span.SetAttributes(attribute.String("subject_id_unprocessed", subjectIdUnprocessed))
	subjectId, err := strconv.Atoi(subjectIdUnprocessed)
	if err != nil {
		return utils.NewParserError(nil, "Invalid subject id (not convertable to int)")
	}
	span.SetAttributes(attribute.Int("subject_id", subjectId))

	roomIdUnprocessed := f.Get("room_id")
	span.SetAttributes(attribute.String("room_id_unprocessed", roomIdUnprocessed))
	roomId, err := strconv.Atoi(roomIdUnprocessed)
	if err != nil {
		return utils.NewParserError(nil, "Invalid room id (not convertable to int)")
	}
	span.SetAttributes(attribute.Int("room_id", roomId))

	schoolIdUnprocessed := f.Get("school_id")
	span.SetAttributes(attribute.String("school_id_unprocessed", schoolIdUnprocessed))
	schoolId, err := strconv.Atoi(schoolIdUnprocessed)
	if err != nil {
		return utils.NewParserError(nil, "Invalid school id (not convertable to int)")
	}
	span.SetAttributes(attribute.Int("school_id", schoolId))

	span.SetAttributes(
		attribute.Int("period id", periodId),
		attribute.Int("subject id", subjectId),
		attribute.Int("room id", roomId),
		attribute.Int("school id", schoolId),
		attribute.String("weekday", weekday),
	)

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
	_, err = tx.Exec(context.TODO(), "insert into regular_timetable (period_id, subject_id, room_id, school_id, weekday) values ($1, $2, $3, $4, $5)", t.periodId, t.subjectId, t.roomId, t.schoolId, t.weekday)
	return
}

func HandleCreateRegularTimetable(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "create regulart timetable")
			defer span.End()

			timetable := reqCtx.Value("regular timetable").(RegularTimetable)

			if err := utils.HandleTx(ctx, db, timetable.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)

		},
	)
}
