package note

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("note")
)

type Note struct {
	id          int
	timetableId int
	noteType    string
	content     string
	date        string
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("parsing note")

	timetableId, err := utils.ParseInt(span, "timetable_id", f.Get("timetable_id"))
	if err != nil {
		return utils.NewParserError(err, "Invalid timetable id (not an int)")
	}
	noteType := f.Get("type")
	span.SetAttributes(
		attribute.String("type", noteType),
	)
	if noteType != "homework" && noteType != "test" {
		return utils.NewParserError(nil, "Invalid note type")
	}
	content := f.Get("content")
	if content == "" {
		return utils.NewParserError(nil, "Content not provided")
	}
	date := ""
	dateUnprocessed := f.Get("date")
	if dateUnprocessed != "" {
		var dateAsTime time.Time
		dateAsTime, err = time.Parse(time.DateOnly, dateUnprocessed)
		date = dateAsTime.Format(time.DateOnly)
	}
	span.SetAttributes(
		attribute.String("date_unprocessed", dateUnprocessed),
		attribute.String("date", date),
	)
	if err != nil {
		return utils.NewParserError(err, "Invalid date")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "note", Note{
		id:          -1,
		timetableId: timetableId,
		noteType:    noteType,
		content:     content,
		date:        date,
	})

	return nil
}

func (n Note) SaveToDB(tx pgx.Tx) (err error) {
	if n.date == "" {
		_, err = tx.Exec(context.TODO(),
			`insert into note (type, content, timetable_id) values ($1, $2, $3)`,
			n.noteType, n.content, n.timetableId,
		)
	} else {
		_, err = tx.Exec(context.TODO(),
			`insert into note_with_date (type, content, timetable_id, date) values ($1, $2, $3, $4)`,
			n.noteType, n.content, n.timetableId, n.date,
		)
	}

	return
}

func HandleCreateNote(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "create note")
		defer span.End()

		note := reqCtx.Value("note").(Note)

		err := utils.HandleTx(ctx, db, note.SaveToDB)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
