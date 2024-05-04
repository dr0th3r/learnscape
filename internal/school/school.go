package school

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/dr0th3r/learnscape/internal/user"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/google/uuid"
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
	id             string
	name           string
	city           string
	zip_code       string
	street_address string
}

func parse(f url.Values, ctx context.Context) (*School, error) {
	span := trace.SpanFromContext(ctx)

	school := School{
		id:             uuid.NewString(),
		name:           f.Get("school_name"),
		city:           f.Get("city"),
		zip_code:       f.Get("zip_code"),
		street_address: f.Get("street_address"),
	}

	span.SetAttributes(
		attribute.String("id", school.id),
		attribute.String("name", school.name),
		attribute.String("city", school.city),
		attribute.String("zip code", school.zip_code),
		attribute.String("street address", school.street_address),
	)

	if school.name == "" || school.city == "" || school.zip_code == "" || school.street_address == "" {
		return nil, errors.New("Missing field(s)")
	}

	return &school, nil
}

func (s *School) saveToDB(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "insert into school (id, name, city, zip_code, street_address) values ($1, $2, $3, $4, $5)",
		s.id, s.name, s.city, s.zip_code, s.street_address,
	)
	if err != nil {
		return err
	}
	return nil
}

func HandleRegisterSchool(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "register school")
			defer span.End()

			span.AddEvent("Starting to parse form data")
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing formdate", ctx)
				return
			}

			span.AddEvent("Starting to parse school from form data")
			school, err := parse(r.Form, ctx)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, err.Error(), ctx)
				return
			}
			span.AddEvent("Starting to parse admin from form data")
			admin, err := user.ParseRegister(r.Form, ctx)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, err.Error(), ctx)
				return
			}

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
