package report

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("report")
)

type Report struct {
	id           int
	timetableId  int
	reportedBy   uuid.UUID
	reportedAt   time.Time
	topicCovered string
}

func ParseReport(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing  report")

	timetableId, err := utils.ParseInt(span, "timetable_id", f.Get("timetable_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid timetable id")
	}
	reportedBy, err := utils.ParseUuid(span, "reported_by", f.Get("reported_by"))
	if err != nil {
		return utils.NewParserError(err, "Invalid reported by field")
	}

	topicCovered := f.Get("topic_covered")
	if topicCovered == "" {
		return utils.NewParserError(nil, "No covered topic was provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, " report", Report{
		id:           -1,
		timetableId:  timetableId,
		reportedBy:   reportedBy,
		reportedAt:   time.Now(),
		topicCovered: topicCovered,
	})

	return nil
}

func (r Report) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(), "insert into report (timetable_id, reported_by, topic_covered) values ($1, $2, $3)",
		r.timetableId, r.reportedBy, r.topicCovered,
	)
	return
}

func HandleCreateReport(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "create report")
		defer span.End()

		Report := ctx.Value(" report").(Report)

		if err := utils.HandleTx(ctx, db, Report.SaveToDB); err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
