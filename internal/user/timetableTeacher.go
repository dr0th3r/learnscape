package user

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerTimetableTeacher = otel.Tracer("timetable teacher")
)

type TimetableTeacher struct {
	timetableId int
	teacherId   uuid.UUID
}

func ParseTimetableTeacher(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing timetable teacher")

	timetableId, err := utils.ParseInt(span, "timetable_id", f.Get("timetable_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid  timetable id (not an int)")
	}
	teacherId, err := utils.ParseUuid(span, "teacher_id", f.Get("teacher_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid teacherId")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "timetable_teacher", TimetableTeacher{
		timetableId: timetableId,
		teacherId:   teacherId,
	})

	return nil
}

func (tt TimetableTeacher) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(),
		"insert into timetable_teacher (timetable_id, teacher_id) values ($1, $2)",
		tt.timetableId, tt.teacherId)
	return
}

func HandleCreateTimetableTeacher(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracerTimetableTeacher.Start(reqCtx, "create regular timetable teacher")
		defer span.End()

		timetableTeacher := reqCtx.Value("timetable_teacher").(TimetableTeacher)

		err := utils.HandleTx(ctx, db, timetableTeacher.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
