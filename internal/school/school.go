package school

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/user"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("school")
)

type School struct {
	id            int
	name          string
	city          string
	zip_code      string
	streetAddress string
}

func Parse(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing school")

	school := School{
		id:            -1,
		name:          f.Get("school_name"),
		city:          f.Get("city"),
		zip_code:      f.Get("zip_code"),
		streetAddress: f.Get("street_address"),
	}

	span.SetAttributes(
		attribute.Int("id", school.id),
		attribute.String("name", school.name),
		attribute.String("city", school.city),
		attribute.String("zip code", school.zip_code),
		attribute.String("street address", school.streetAddress),
	)

	if school.name == "" {
		return utils.NewParserError(nil, "Name not provided")
	} else if school.city == "" {
		return utils.NewParserError(nil, "City not provided")
	} else if school.zip_code == "" {
		return utils.NewParserError(nil, "Zip code not provided")
	} else if school.streetAddress == "" {
		return utils.NewParserError(nil, "Street address not provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "school", school)

	return nil
}

func (s School) saveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "insert into school (name, city, zip_code, street_address) values ($1, $2, $3, $4)",
		s.name, s.city, s.zip_code, s.streetAddress,
	)
	if err != nil {
		return err
	}
	return nil
}

func HandleRegisterSchool(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			ctx, span := tracer.Start(reqCtx, "register school")
			defer span.End()

			school := reqCtx.Value("school").(School)
			admin := reqCtx.Value("user").(user.User)

			if err := utils.HandleTx(ctx, db, school.saveToDB, admin.SaveToDB); err != nil {
				utils.UnexpectedError(w, err, ctx)
			}

			span.AddEvent("Starting to set jwt token for admin")
			if err := admin.SetToken(w, []byte("my secret"), time.Now().Add(time.Hour*72)); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error setting jwt", ctx)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
