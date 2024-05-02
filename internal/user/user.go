package user

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dr0th3r/learnscape/internal/utils"
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

func parseRegister(f url.Values) (*User, error) {
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

func parseLogin(f url.Values) (*User, error) {
	user := User{
		email:    f.Get("email"),
		password: f.Get("password"),
	}

	if user.email == "" || user.password == "" {
		return nil, errors.New("Missing field(s)")
	}

	return &user, nil
}

func (u *User) saveToDB(db *pgxpool.Pool, rdb *redis.Client) (*string, error) {
	password_hash, err := argon2id.CreateHash(u.password, argon2id.DefaultParams)

	ctx := context.Background()

	_, err = db.Exec(ctx, "insert into users (id, name, surname, email, password) values ($1, $2, $3, $4, $5)",
		u.id, u.name, u.surname, u.email, password_hash)
	if err != nil {
		return nil, err
	}

	sessionId := uuid.NewString()
	if err := rdb.Set(ctx, "session_id", sessionId, time.Hour*72).Err(); err != nil {
		return nil, err
	}
	return &sessionId, nil
}

func (u *User) login(db *pgxpool.Pool, rdb *redis.Client) (*string, error) {
	ctx := context.Background()

	var dbPassword string
	if err := db.QueryRow(ctx, "select password from users where email=$1", u.email).Scan(&dbPassword); err != nil {
		return nil, err
	}

	passwordsMatch, err := argon2id.ComparePasswordAndHash(u.password, dbPassword)
	if err != nil {
		return nil, err
	}
	if !passwordsMatch {
		return nil, errors.New("Passwords do not match")
	}

	sessionId := uuid.NewString()
	if err := rdb.Set(ctx, "session_id", sessionId, time.Hour*72).Err(); err != nil {
		return nil, err
	}
	return &sessionId, nil
}

func setSessionId(w http.ResponseWriter, sessionId *string) error {
	if sessionId == nil {
		return errors.New("Session id was not provided")
	}

	cookie := http.Cookie{
		Name:     "sessionId",
		Value:    *sessionId,
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	return nil
}

func HandleRegisterUser(db *pgxpool.Pool, rdb *redis.Client) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing form data")
				return
			}

			user, err := parseRegister(r.Form)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, "")
				return
			}

			sessionId, err := user.saveToDB(db, rdb)
			if err != nil {
				var code int
				var msg string

				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					code = http.StatusConflict
					msg = "Email already registered"
				} else {
					code = http.StatusInternalServerError
					msg = "Failed to save user to db"
				}

				utils.HandleError(w, err, code, msg)
				return
			}

			if err := setSessionId(w, sessionId); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Failed to set session id")
			}
		},
	)
}

func HandleLogin(db *pgxpool.Pool, rdb *redis.Client) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing form data")
				return
			}

			user, err := parseLogin(r.Form)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, "")
				return
			}

			sessionId, err := user.login(db, rdb)
			if err != nil {
				utils.HandleError(w, err, http.StatusUnauthorized, "Failed to log user in")
				return
			}

			if err := setSessionId(w, sessionId); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Failed to set session id")
			}
		},
	)
}
