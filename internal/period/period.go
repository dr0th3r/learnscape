package period

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("period")
)

type Period struct {
	id       int
	schoolId int
	start    string
	end      string
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing period")

	schoolId, err := utils.ParseInt(span, "school_id", f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid school id")
	}

	start := f.Get("start")
	_, err = time.Parse(time.TimeOnly, start)
	span.SetAttributes(
		attribute.String("start_unprocessed", start),
		attribute.String("start", start),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid start time")
	}

	end := f.Get("end")
	_, err = time.Parse(time.TimeOnly, end)
	span.SetAttributes(
		attribute.String("end_unprocessed", end),
		attribute.String("end", end),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid end time")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "period", Period{
		id:       -1,
		schoolId: schoolId,
		start:    start,
		end:      end,
	})

	return nil
}

func (p Period) SaveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "insert into period (school_id, span) values($1, $2)",
		p.schoolId, fmt.Sprintf("[%s, %s]", p.start, p.end),
	)
	if err != nil {
		return err
	}

	return nil
}

func HandleCreatePeriod(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "period creation")
			defer span.End()

			period := reqCtx.Value("period").(Period)

			if err := utils.HandleTx(ctx, db, period.SaveToDB); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "22000" {
					utils.HandleError(w, err, http.StatusBadRequest, "Period times overlap or start is before end", ctx)
				} else {
					utils.UnexpectedError(w, err, ctx)
				}

				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
