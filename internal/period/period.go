package period

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
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
	id        int
	school_id uuid.UUID
	start     string
	end       string
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing period")

	school_id, err := uuid.Parse(f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid school id")
	}
	span.SetAttributes(attribute.String("school_id", school_id.String()))

	start := f.Get("start")
	_, err = time.Parse(time.TimeOnly, start)
	if err != nil {
		return utils.NewParserError(err, "Invalid start time")
	}
	span.SetAttributes(attribute.String("start", start))

	end := f.Get("end")
	_, err = time.Parse(time.TimeOnly, end)
	if err != nil {
		return utils.NewParserError(err, "Invalid end time")
	}
	span.SetAttributes(attribute.String("end", end))

	*handlerCtx = context.WithValue(*handlerCtx, "period", Period{
		id:        -1,
		school_id: school_id,
		start:     start,
		end:       end,
	})

	return nil
}

func (p Period) SaveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "insert into period (school_id, span) values($1, $2)",
		p.school_id, fmt.Sprintf("[%s, %s]", p.start, p.end),
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
