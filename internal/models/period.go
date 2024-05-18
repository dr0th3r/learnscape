package models

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
)

type Period struct {
	id       int
	schoolId int
	start    time.Time
	end      time.Time
}

func ParsePeriod(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing period")

	start, err := utils.ParseTime(span, "start", f.Get("start"), time.TimeOnly)
	if err != nil {
		return utils.NewParserError(err, "Invalid start time")
	}
	end, err := utils.ParseTime(span, "end", f.Get("end"), time.TimeOnly)
	if err != nil {
		return utils.NewParserError(err, "Invalid end time")
	}

	if end.Before(start) {
		return utils.NewParserError(nil, "End can't be before start")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "period", Period{
		id:       -1,
		schoolId: -1,
		start:    start,
		end:      end,
	})

	return nil
}

func (p Period) SaveToDBWithSchoolId(schoolId int) utils.TxFunc {
	return func(tx pgx.Tx) error {
		_, err := tx.Exec(context.Background(), "insert into period (school_id, span) values($1, $2)",
			schoolId,
			fmt.Sprintf(
				"[%s, %s]",
				p.start.Format(time.TimeOnly),
				p.end.Format(time.TimeOnly),
			),
		)
		if err != nil {
			return err
		}

		return nil

	}
}
