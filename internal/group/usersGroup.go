package group

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerUsersGroup = otel.Tracer("users group")
)

type UsersGroup struct {
	userId  uuid.UUID
	groupId int
}

func ParseUsersGroup(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing users group")

	groupId, err := utils.ParseInt(span, "group_id", f.Get("group_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid group id (not an int)")
	}
	userId, err := utils.ParseUuid(span, "user_id", f.Get("user_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid user id")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "users_group", UsersGroup{
		userId:  userId,
		groupId: groupId,
	})

	return nil
}

func (ug UsersGroup) SaveToDB(tx pgx.Tx) (err error) {
	_, err = tx.Exec(context.TODO(),
		"insert into users_group (user_id, group_id) values ($1, $2)",
		ug.userId, ug.groupId)
	return
}

func HandleCreateUsersGroup(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracerUsersGroup.Start(reqCtx, "create users group")
		defer span.End()

		usersGroup := reqCtx.Value("users_group").(UsersGroup)

		err := utils.HandleTx(ctx, db, usersGroup.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
