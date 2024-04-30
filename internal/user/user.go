package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	id       string
	name     string
	surname  string
	email    string
	password string
}

func parseFromForm(f url.Values) (*User, error) {
	user := User{
		id:       uuid.NewString(),
		name:     f.Get("name"),
		surname:  f.Get("surname"),
		email:    f.Get("email"),
		password: f.Get("password"),
	}

	if user.name == "" || user.surname == "" || user.email == "" || user.password == "" {
		return nil, errors.New("Missing field(s)")
	}

	return &user, nil
}

func (u *User) saveToDB(db *pgxpool.Pool) error {
	password_hash, err := argon2id.CreateHash(u.password, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), "insert into users (id, name, surname, email, password) values ($1, $2, $3, $4, $5)",
		u.id, u.name, u.surname, u.email, password_hash)
	return err
}

func HandleRegisterUser(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				fmt.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error parsing form data")
				return
			}

			user, err := parseFromForm(r.Form)
			if err != nil {
				fmt.Println(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, err.Error())
				return
			}

			if err := user.saveToDB(db); err != nil {
				fmt.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Failed to save user to db")
			}
		},
	)
}