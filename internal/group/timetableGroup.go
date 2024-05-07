package group

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerTimetableGroup = otel.Tracer(" timetable group")
)

type TimetableGroup struct {
	timetableId int
	groupId     int
}

func ParseTimetableGroup(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing timetable group")

	timetableId, err := utils.ParseInt(span, "timetable_id", f.Get("timetable_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid timetable id (not an int)")
	}
	groupId, err := utils.ParseInt(span, "group_id", f.Get("group_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid group id (not an int)")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "timetable_group", TimetableGroup{
		timetableId: timetableId,
		groupId:     groupId,
	})

	return nil
}

func (tg TimetableGroup) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(),
		"insert into timetable_group (timetable_id, group_id) values ($1, $2)",
		tg.timetableId, tg.groupId)
	return
}

func HandleCreateTimetableGroup(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracerUsersGroup.Start(reqCtx, "create timetable group")
		defer span.End()

		usersGroup := reqCtx.Value("timetable_group").(TimetableGroup)

		err := utils.HandleTx(ctx, db, usersGroup.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
