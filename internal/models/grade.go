package models

import (
	"context"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
)

type Grade struct {
	id        int
	studentId uuid.UUID
	reportId  int
	value     int
	weight    int
}

func ParseGrade(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing grade")

	studentId, err := utils.ParseUuid(span, "student_id", f.Get("student_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid student id")
	}
	reportId, err := utils.ParseInt(span, "report_id", f.Get("report_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid report id (not an int)")
	}
	value, err := utils.ParseInt(span, "value", f.Get("value"))
	if err != nil {
		return utils.NewParserError(err, "Invalid grade value (not an int)")
	} else if value < 1 {
		return utils.NewParserError(nil, "Invalid grade value (can't be less than 1)")
	} else if value > 5 {
		return utils.NewParserError(nil, "Invalid grade value (can't be more than 5)")
	}
	weight, err := utils.ParseInt(span, "weight", f.Get("weight"))
	if err != nil {
		return utils.NewParserError(err, "Invalid grade weight (not an int)")
	} else if weight < 1 {
		return utils.NewParserError(nil, "Invalid grade weight (can't be less than 1)")
	} else if weight > 10 {
		return utils.NewParserError(nil, "Invalid grade weight (can't be more than 10)")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "grade", Grade{
		id:        -1,
		studentId: studentId,
		reportId:  reportId,
		value:     value,
		weight:    weight,
	})

	return nil
}

func (g Grade) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(), "insert into grade (student_id, report_id, value, weight) values ($1, $2, $3, $4)", g.studentId, g.reportId, g.value, g.weight)

	return
}
