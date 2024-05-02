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
)

type School struct {
	id             string
	name           string
	city           string
	zip_code       string
	street_address string
}

func parse(f url.Values) (*School, error) {
	school := School{
		id:             uuid.NewString(),
		name:           f.Get("school_name"),
		city:           f.Get("city"),
		zip_code:       f.Get("zip_code"),
		street_address: f.Get("street_address"),
	}

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
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing formdate")
				return
			}

			school, err := parse(r.Form)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, err.Error())
				return
			}
			admin, err := user.ParseRegister(r.Form)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, err.Error())
				return
			}

			tx, err := db.Begin(context.Background())
			defer tx.Rollback(context.Background())
			if err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "")
			}
			if err := school.saveToDB(tx); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "")
			}
			if err := admin.SaveToDB(tx); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "")
			}
			if err := tx.Commit(context.Background()); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "")
			}

			if err := admin.SetToken(w, []byte("my secret"), time.Now().Add(time.Hour*72)); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error setting jwt")
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}
