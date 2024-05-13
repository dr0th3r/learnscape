package models

import (
	"context"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
)

type ParentChild struct {
	parentId uuid.UUID
	childId  uuid.UUID
}

func ParseParentChild(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing user")

	parentId, err := utils.ParseUuid(span, "parent_id", f.Get("parent_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid parent id")
	}
	childId, err := utils.ParseUuid(span, "child_id", f.Get("child_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid child id")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "parent child", ParentChild{
		parentId: parentId,
		childId:  childId,
	})

	return nil
}

func (pc ParentChild) SaveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.TODO(), "insert into parent_child (parent_id, child_id) values ($1, $2)",
		pc.parentId.String(), pc.childId.String(),
	)
	if err != nil {
		return err
	}

	return nil
}
