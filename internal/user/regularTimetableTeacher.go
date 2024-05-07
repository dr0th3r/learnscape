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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerRegularTimetableTeacher = otel.Tracer("regualar timetable teacher")
)

type RegularTimetableTeacher struct {
	regularTimetableId int
	teacherId          uuid.UUID
}

func ParseRegularTimetableTeacher(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing regular timetable teacher")

	regularTimetableId, err := utils.ParseInt(span, "regular_timetable_id", f.Get("regular_timetable_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid regular timetable id (not an int)")
	}
	teacherIdUnprocessed := f.Get("teacher_id")
	teacherId, err := uuid.Parse(teacherIdUnprocessed)
	span.SetAttributes(
		attribute.String("teacher_id_unprocessed", teacherIdUnprocessed),
		attribute.String("teacher_id", teacherId.String()),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid teacherId")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "regular_timetable_teacher", RegularTimetableTeacher{
		regularTimetableId: regularTimetableId,
		teacherId:          teacherId,
	})

	return nil
}

func (rtt RegularTimetableTeacher) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(),
		"insert into regular_timetable_teacher (regular_timetable_id, teacher_id) values ($1, $2)",
		rtt.regularTimetableId, rtt.teacherId)
	return
}

func HandleCreateRegularTimetableTeacher(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracerRegularTimetableTeacher.Start(reqCtx, "create regular timetable teacher")
		defer span.End()

		timetableTeacher := reqCtx.Value("regular_timetable_teacher").(RegularTimetableTeacher)

		err := utils.HandleTx(ctx, db, timetableTeacher.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
