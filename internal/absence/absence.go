package absence

import (
	"context"
	"fmt"
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
	tracer = otel.Tracer("absence")
)

type Absence struct {
	userId uuid.UUID
	start  time.Time
	end    time.Time
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing absence")

	userId, err := utils.ParseUuid(span, "user_id", f.Get("user_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid user id")
	}
	start, err := utils.ParseTime(span, "start", f.Get("start"), time.RFC3339)
	if err != nil {
		return utils.NewParserError(err, "Invalid start time")
	}
	end, err := utils.ParseTime(span, "end", f.Get("end"), time.RFC3339)
	if err != nil {
		return utils.NewParserError(err, "Invalid end time")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "absence", Absence{
		userId: userId,
		start:  start,
		end:    end,
	})

	return nil
}

func (a Absence) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(),
		"insert into absence (user_id, span) values ($1, $2)",
		a.userId, fmt.Sprintf(
			"[%s, %s]",
			a.start.Format(time.RFC3339),
			a.end.Format(time.RFC3339),
		),
	)
	return
}

func HandleCreateAbsence(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "create absence")
		defer span.End()

		absence := reqCtx.Value("absence").(Absence)

		err := utils.HandleTx(ctx, db, absence.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
