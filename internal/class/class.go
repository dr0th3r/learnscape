package class

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
	tracer = otel.Tracer("class")
)

type Class struct {
	id             int
	name           string
	year           int8
	classTeacherId uuid.UUID
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing class")

	year, err := utils.ParseInt(span, "year", f.Get("year"))
	if err != nil {
		return utils.NewParserError(err, "Invalid year (not an integer)")
	} else if year > 9 {
		return utils.NewParserError(nil, "Invalid year (too high)")
	} else if year <= 0 {
		return utils.NewParserError(nil, "Invalid year (can't be 0 or less)")
	}

	classTeacherId, err := utils.ParseUuid(span, "class_teacher_id", f.Get("class_teacher_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid class teacher id")
	}

	class := Class{
		id:             -1,
		name:           f.Get("name"),
		year:           int8(year),
		classTeacherId: classTeacherId,
	}

	if class.name == "" {
		return utils.NewParserError(nil, "Name not provided")
	} else {
		span.SetAttributes(attribute.String("name", class.name))
	}

	*handlerCtx = context.WithValue(*handlerCtx, "class", class)

	return nil
}

func (c Class) SaveToDb(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(),
		"insert into class (name, year, class_teacher_id) values ($1, $2, $3)",
		c.name, c.year, c.classTeacherId,
	)
	return
}

func HandleCreateClass(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "create class")
		defer span.End()

		class := reqCtx.Value("class").(Class)

		if err := utils.HandleTx(ctx, db, class.SaveToDb); err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
