package subject

import (
	"context"
	"errors"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func (s Subject) SaveToDB(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), "insert into subject (name, mandatory) values ($1, $2)", s.name, s.mandatory)
	if err != nil {
		return err
	}
	return nil
}
