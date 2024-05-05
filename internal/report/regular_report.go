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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("regular report")
)

type RegularReport struct {
	id                 int
	regularTimetableId int
	reportedBy         uuid.UUID
	reportedAt         time.Time
	topicCovered       string
}

func ParseRegularReport(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing regular report")

	regularTimetableId, err := utils.ParseInt(span, "regular_timetable_id", f.Get("regular_timetable_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid regular timetable id")
	}
	reportedByUnprocessed := f.Get("reported_by")
	reportedBy, err := uuid.Parse(reportedByUnprocessed)
	span.SetAttributes(
		attribute.String("reported_by_unprocessed", reportedByUnprocessed),
		attribute.String("reported_by", reportedBy.String()),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid reported by field")
	}

	topicCovered := f.Get("topic_covered")
	if topicCovered == "" {
		return utils.NewParserError(nil, "No covered topic was provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "regular report", RegularReport{
		id:                 -1,
		regularTimetableId: regularTimetableId,
		reportedBy:         reportedBy,
		reportedAt:         time.Now(),
		topicCovered:       topicCovered,
	})

	return nil
}

func (r RegularReport) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(), "insert into regular_report (regular_timetable_id, reported_by, topic_covered) values ($1, $2, $3)",
		r.regularTimetableId, r.reportedBy, r.topicCovered,
	)
	return
}

func HandleCreateRegularReport(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "create regular report")
		defer span.End()

		regularReport := ctx.Value("regular report").(RegularReport)

		if err := utils.HandleTx(ctx, db, regularReport.SaveToDB); err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
