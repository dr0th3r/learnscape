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

func Parse(f url.Values, ctx context.Context) (Period, error) {
	span := trace.SpanFromContext(ctx)

	school_id, err := uuid.Parse(f.Get("school_id"))
	if err != nil {
		return Period{}, err
	}
	span.SetAttributes(attribute.String("school_id", school_id.String()))

	start := f.Get("start")
	_, err = time.Parse(time.TimeOnly, start)
	if err != nil {
		return Period{}, err
	}
	span.SetAttributes(attribute.String("start", start))

	end := f.Get("end")
	_, err = time.Parse(time.TimeOnly, end)
	if err != nil {
		return Period{}, err
	}
	span.SetAttributes(attribute.String("end", end))

	return Period{
		id:        -1,
		school_id: school_id,
		start:     start,
		end:       end,
	}, nil
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
			ctx, span := tracer.Start(r.Context(), "period creation")
			defer span.End()

			span.AddEvent("Starting to parse form data")
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing form data", ctx)
				return
			}

			span.AddEvent("Starting to parse period from form data")
			period, err := Parse(r.Form, ctx)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, "Invalid period", ctx)
				return
			}

			span.AddEvent("Starting database transaction")
			tx, err := db.Begin(context.Background())
			if err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "", ctx)
				return
			}
			defer tx.Rollback(context.Background())
			span.AddEvent("Starting to save period to database")
			if err := period.SaveToDB(tx); err != nil {
				var code int
				var msg string

				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) {
					code = http.StatusBadRequest
					msg = "Data is invalid" //TODO: add better error messages later
				} else {
					code = http.StatusInternalServerError
					msg = "Internal server error"
				}

				utils.HandleError(w, err, code, msg, ctx)
				return
			}
			span.AddEvent("Starting to commit database transaction")
			if err := tx.Commit(context.Background()); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "", ctx)
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
