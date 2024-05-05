package subject

import (
	"context"
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
	schoolId  int
	name      string
	mandatory bool
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing subject")

	schoolId, err := utils.ParseInt(span, "school_id", f.Get("school_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid school id")
	}

	mandatory := true
	if f.Get("mandatory") == "false" {
		mandatory = false
	}

	subject := Subject{
		id:        -1,
		schoolId:  schoolId,
		name:      f.Get("name"),
		mandatory: mandatory,
	}
	span.SetAttributes(
		attribute.String("name", subject.name),
		attribute.Bool("mandatory", subject.mandatory),
	)

	if subject.name == "" {
		return utils.NewParserError(nil, "Subject name not provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "subject", subject)

	return nil
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
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "subject creation")
			defer span.End()

			subject := reqCtx.Value("subject").(Subject)

			if err := utils.HandleTx(ctx, db, subject.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
