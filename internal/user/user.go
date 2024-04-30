package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

func (u *User) saveToDB(db *pgxpool.Pool, rdb *redis.Client) (string, error) {
	password_hash, err := argon2id.CreateHash(u.password, argon2id.DefaultParams)

	ctx := context.Background()

	_, err = db.Exec(ctx, "insert into users (id, name, surname, email, password) values ($1, $2, $3, $4, $5)",
		u.id, u.name, u.surname, u.email, password_hash)
	if err != nil {
		return "", err
	}

	sessionId := uuid.NewString()
	if err := rdb.Set(ctx, "session_id", sessionId, time.Hour*72).Err(); err != nil {
		return "", err
	}
	return sessionId, nil
}

func HandleRegisterUser(db *pgxpool.Pool, rdb *redis.Client) http.Handler {
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

			sessionId, err := user.saveToDB(db, rdb)
			if err != nil {
				fmt.Println(err.Error())

				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					w.WriteHeader(http.StatusConflict)
					fmt.Fprintf(w, "Email already registered")
					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Failed to save user to db")
			}

			cookie := http.Cookie{
				Name:     "sessionId",
				Value:    sessionId,
				Expires:  time.Now().Add(72 * time.Hour),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}

			http.SetCookie(w, &cookie)
		},
	)
}
