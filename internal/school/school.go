package school

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/dr0th3r/learnscape/internal/user"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type School struct {
	id             int
	name           string
	city           string
	zip_code       string
	street_address string
}

func parse(f url.Values) (*School, error) {
	school := School{
		id:             -1,
		name:           f.Get("name"),
		city:           f.Get("city"),
		zip_code:       f.Get("zip_code"),
		street_address: f.Get("street_address"),
	}

	if school.name == "" || school.city == "" || school.zip_code == "" || school.street_address == "" {
		return nil, errors.New("Missing field(s)")
	}

	return &school, nil
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

			_ = school
			_ = admin
		},
	)
}
