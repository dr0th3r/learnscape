package subject

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("subject")
)

type Subject struct {
	id        int
	name      string
	mandatory bool
}

func Parse(f url.Values, ctx context.Context) (Subject, error) {
	span := trace.SpanFromContext(ctx)

	mandatory := true
	if f.Get("mandatory") == "false" {
		mandatory = false
	}

	subject := Subject{
		id:        -1,
		name:      f.Get("name"),
		mandatory: mandatory,
	}
	span.SetAttributes(
		attribute.String("name", subject.name),
		attribute.Bool("mandatory", subject.mandatory),
	)

	if subject.name == "" {
		return Subject{}, errors.New("Missing field(s)")
	}

	return subject, nil
}

func (s Subject) SaveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "insert into subject (name, mandatory) values ($1, $2)", s.name, s.mandatory)
	if err != nil {
		return err
	}
	return nil
}

func HandleCreateSubject(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "subject creation")
			defer span.End()

			span.AddEvent("Starting to parse form data")
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing form data", ctx)
				return
			}

			span.AddEvent("Starting to parse subject from form data")
			subject, err := Parse(r.Form, ctx)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, "Invalid subject", ctx)
				return
			}

			if err := utils.HandleTx(ctx, db, subject.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
