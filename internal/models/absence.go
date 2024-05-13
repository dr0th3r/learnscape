package models

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
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
