package group

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
	tracer = otel.Tracer("group")
)

type Group struct {
	id      int
	classId int
	name    string
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
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

func HandleCreateGroup(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "create group")
		defer span.End()

		group := reqCtx.Value("group").(Group)

		err := utils.HandleTx(ctx, db, group.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
