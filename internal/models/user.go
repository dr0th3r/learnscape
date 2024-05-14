package models

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
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func ParseRegister(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing user")

	emailUnprocessed := f.Get("email")
	span.SetAttributes(attribute.String("email_unprocessed", emailUnprocessed))
	email, err := mail.ParseAddress(emailUnprocessed)
	if err != nil {
		return utils.NewParserError(err, "Invalid email provided")
	}
	span.SetAttributes(attribute.String("email", email.Address))

	password := f.Get("password")
	if err := validatePassword(password); err != nil {
		return utils.NewParserError(err, "Invalid password provided")
	}

	user := User{
		id:       uuid.NewString(),
		name:     f.Get("user_name"),
		surname:  f.Get("surname"),
		email:    email.Address,
		password: password,
	}

	span.SetAttributes(
		attribute.String("id", user.id),
		attribute.String("name", user.name),
		attribute.String("surname", user.surname),
	)

	if user.name == "" {
		return utils.NewParserError(nil, "User name not provided") //nil means use msg as error
	}

	if user.name == "" || user.surname == "" || user.email == "" || user.password == "" {
		return utils.NewParserError(nil, "Surname not provided")
	}

	*handlerCtx = context.WithValue(*handlerCtx, "user", user)

	return nil
}

func ParseLogin(f url.Values, parserCtx context.Context, handlerCtx *context.Context) *utils.ParseError {
	span := trace.SpanFromContext(parserCtx)
	span.AddEvent("Parsing user")

	email, err := mail.ParseAddress(f.Get("email"))
	if err != nil {
		return utils.NewParserError(err, "Invalid email provided")
	}
	password := f.Get("password")
	if err := validatePassword(password); err != nil {
		return utils.NewParserError(err, "Invalid password provided")
	}

	user := User{
		email:    email.Address,
		password: password,
	}
	span.SetAttributes(
		attribute.String("email", user.email),
	)

	*handlerCtx = context.WithValue(*handlerCtx, "user", user)

	return nil
}

func (u User) SaveToDB(tx pgx.Tx) error {
	password_hash, err := argon2id.CreateHash(u.password, argon2id.DefaultParams)

	_, err = tx.Exec(context.Background(), "insert into users (id, name, surname, email, password) values ($1, $2, $3, $4, $5)",
		u.id, u.name, u.surname, u.email, password_hash)
	if err != nil {
		return err
	}

	return nil
}

func (u User) Login(db *pgxpool.Pool) error {
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

func (u User) SetToken(w http.ResponseWriter, secret []byte, exp time.Time) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		/*jwt.MapClaims{
			"id":      u.id,
			"name":    u.name,
			"surname": u.surname,
			"email":   u.email,
			"exp":     exp.Unix(),
		})*/
		utils.UserClaims{
			Id:      u.id,
			Name:    u.name,
			Surname: u.surname,
			Email:   u.email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(exp),
			},
		},
	)

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
