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

	periodId, err := strconv.Atoi(f.Get("period_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid period id (not convertable to int)")
	}
	subjectId, err := strconv.Atoi(f.Get("subject_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid subject id (not convertable to int)")
	}
	roomId, err := strconv.Atoi(f.Get("room_id"))
	if err != nil {
		return utils.NewParserError(nil, "Invalid room id (not convertable to int)")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "regular timetable", RegularTimetable{
		id:        -1,
		periodId:  periodId,
		subjectId: subjectId,
		roomId:    roomId,
		weekday:   weekday,
	})

	return nil
}

func (t RegularTimetable) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(), "insert into regular_timetable (period_id, subject_id, room_id, weekday) values ($1, $2, $3, $4)", t.periodId, t.subjectId, t.roomId, t.weekday)
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
