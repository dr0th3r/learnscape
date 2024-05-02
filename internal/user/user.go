package user

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	id       string
	name     string
	surname  string
	email    string
	password string
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("Password is too short")
	}

	return nil
}

func ParseRegister(f url.Values) (*User, error) {
	email, err := mail.ParseAddress(f.Get("email"))
	if err != nil {
		return nil, err
	}
	password := f.Get("password")
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	user := User{
		id:       uuid.NewString(),
		name:     f.Get("name"),
		surname:  f.Get("surname"),
		email:    email.Address,
		password: password,
	}

	if user.name == "" || user.surname == "" || user.email == "" || user.password == "" {
		return nil, errors.New("Missing field(s)")
	}

	return &user, nil
}

func parseLogin(f url.Values) (*User, error) {
	email, err := mail.ParseAddress(f.Get("email"))
	if err != nil {
		return nil, err
	}
	password := f.Get("password")
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	user := User{
		email:    email.Address,
		password: password,
	}

	if user.email == "" || user.password == "" {
		return nil, errors.New("Missing field(s)")
	}

	return &user, nil
}

func (u *User) SaveToDB(tx pgx.Tx) error {
	password_hash, err := argon2id.CreateHash(u.password, argon2id.DefaultParams)

	_, err = tx.Exec(context.Background(), "insert into users (id, name, surname, email, password) values ($1, $2, $3, $4, $5)",
		u.id, u.name, u.surname, u.email, password_hash)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) login(db *pgxpool.Pool) error {
	ctx := context.Background()

	var dbPassword string
	if err := db.QueryRow(ctx, "select password from users where email=$1", u.email).Scan(&dbPassword); err != nil {
		return err
	}

	passwordsMatch, err := argon2id.ComparePasswordAndHash(u.password, dbPassword)
	if err != nil {
		return err
	}
	if !passwordsMatch {
		return errors.New("Passwords do not match")
	}

	return nil
}

func (u *User) SetToken(w http.ResponseWriter, secret []byte, exp time.Time) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":      u.id,
			"name":    u.name,
			"surname": u.surname,
			"email":   u.email,
			"exp":     exp.Unix(),
		})

	tokentStr, err := token.SignedString(secret)
	if err != nil {
		return err
	}

	tokenCookie := http.Cookie{
		Name:     "token",
		Value:    tokentStr,
		Expires:  exp,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode, //TODO: add other config by OWASP later
	}
	http.SetCookie(w, &tokenCookie)

	return nil
}

func HandleRegisterUser(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error parsing form data")
				return
			}

			user, err := ParseRegister(r.Form)
			if err != nil {
				utils.HandleError(w, err, http.StatusBadRequest, "")
				return
			}

			tx, err := db.Begin(context.Background()) //tx is necessary becuase this is not the only endpoint using SveToDB
			if err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "")
			}
			defer tx.Rollback(context.Background())
			if err := user.SaveToDB(tx); err != nil {
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
			if err := tx.Commit(context.Background()); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "")
			}

			if err := user.SetToken(w, []byte("my secret"), time.Now().Add(time.Hour*72)); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error setting jwt")
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}

func HandleLogin(db *pgxpool.Pool) http.Handler {
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

			if err := user.login(db); err != nil {
				utils.HandleError(w, err, http.StatusUnauthorized, "Failed to log user in")
				return
			}

			if err := user.SetToken(w, []byte("my secret"), time.Now().Add(time.Hour*72)); err != nil {
				utils.HandleError(w, err, http.StatusInternalServerError, "Error setting jwt")
			}
		},
	)
}
