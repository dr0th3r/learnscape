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
	tracerRegularTimetableGroup = otel.Tracer("regular timetable group")
)

type RegularTimetableGroup struct {
	regularTimetableId int
	groupId            int
}

func ParseRegularTimetableGroup(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing regular timetable group")

	regularTimetableId, err := utils.ParseInt(span, "regular_timetable_id", f.Get("regular_timetable_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid regular timetable id (not an int)")
	}
	groupId, err := utils.ParseInt(span, "group_id", f.Get("group_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid group id (not an int)")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "regular_timetable_group", RegularTimetableGroup{
		regularTimetableId: regularTimetableId,
		groupId:            groupId,
	})

	return nil
}

func (rtg RegularTimetableGroup) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(),
		"insert into regular_timetable_group (regular_timetable_id, group_id) values ($1, $2)",
		rtg.regularTimetableId, rtg.groupId)
	return
}

func HandleCreateRegularTimetableGroup(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracerUsersGroup.Start(reqCtx, "create regular timetable group")
		defer span.End()

		usersGroup := reqCtx.Value("regular_timetable_group").(RegularTimetableGroup)

		err := utils.HandleTx(ctx, db, usersGroup.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
