package models

import (
	"context"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Group struct {
	id      int
	classId int
	name    string
}

func ParseGroup(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing group")

	classId, err := utils.ParseInt(span, "class_id", f.Get("class_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid class id (not an int)")
	}

	group := Group{
		id:      -1,
		classId: classId,
		name:    f.Get("name"),
	}

	span.SetAttributes(attribute.String("name", group.name))
	if group.name == "" {
		return utils.NewParserError(nil, "Group name not provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "group", group)

	return nil
}

func (g Group) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(), `insert into "group" (name, class_id) values ($1, $2)`, g.name, g.classId)
	return
}
