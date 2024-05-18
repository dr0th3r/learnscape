package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"testing"

	i "github.com/dr0th3r/learnscape/internal"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
)

func TestUser(t *testing.T) {
	config, err := utils.ParseConfig()
	if err != nil {
		t.Error(err)
	}

	connectionUrl := config.DB.GetConnectionUrlWithoutName()
	db_name := "test_" + fmt.Sprint(rand.Int())
	config.DB.Name = db_name

	if err := createNewDB(connectionUrl, db_name); err != nil {
		t.Error(err)
		return
	}

	t.Cleanup(func() {
		if err := dropDB(connectionUrl, db_name); err != nil {
			fmt.Println(err)
		}
	})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	go i.Run(ctx, config)

	if err := waitForReady(ctx); err != nil {
		t.Error(err)
	}
	conn, err := pgx.Connect(context.Background(), config.DB.GetConnectionUrl())
	if err != nil {
		t.Error(err)
	}
	schoolId, err := createSchool(conn)
	if err != nil {
		t.Error(err)
	}

	schoolIdStr := fmt.Sprint(schoolId)

	register_url := "http://localhost:8080/register_user"
	login_url := "http://localhost:8080/login"

	t.Run("incomplete body returns 400 bad request", func(t *testing.T) {
		res, err := http.PostForm(register_url, url.Values{
			"user_name": {"test"},
			"surname":   {"idk"},
			"email":     {"test@idk.com"},
			//password is missing
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		got := res.StatusCode
		want := http.StatusBadRequest
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}
	})

	t.Run("invalid email is rejected", func(t *testing.T) {
		res, err := http.PostForm(register_url, url.Values{
			"user_name": {"test"},
			"surname":   {"idk"},
			"email":     {"invalid"},
			"password":  {"test123456"},
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		got := res.StatusCode
		want := http.StatusBadRequest
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}
	})

	t.Run("invalid password is rejected", func(t *testing.T) {
		res, err := http.PostForm(register_url, url.Values{
			"user_name": {"test"},
			"surname":   {"idk"},
			"email":     {"random@email.com"},
			"password":  {"123"}, //password is too short
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		got := res.StatusCode
		want := http.StatusBadRequest
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}
	})

	t.Run("user without school_id is rejected", func(t *testing.T) {
		res, err := http.PostForm(register_url, url.Values{
			"user_name": {"test"},
			"surname":   {"idk"},
			"email":     {"random@email.com"},
			"password":  {"test123456"},
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		got := res.StatusCode
		want := http.StatusBadRequest
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}
	})

	t.Run("valid user is created", func(t *testing.T) {
		res, err := http.PostForm(register_url, url.Values{
			"user_name": {"test"},
			"surname":   {"idk"},
			"email":     {"random2@email.com"},
			"school_id": {schoolIdStr},
			"password":  {"test123456"},
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		gotCode := res.StatusCode
		wantCode := http.StatusCreated
		if gotCode != wantCode {
			t.Errorf("Got %d, want %d", gotCode, wantCode)
		}

		gotCookies := res.Cookies()

		gotCookiesLen := len(res.Cookies())
		wantCookiesLen := 1
		if gotCookiesLen != wantCookiesLen {
			t.Errorf("Got %d cookies, wanted %d", gotCookiesLen, wantCookiesLen)
		}

		if gotCookies[0].Name != "token" || !gotCookies[0].HttpOnly || gotCookies[0].SameSite != http.SameSiteStrictMode {
			t.Errorf("Got %s invalid cookie", gotCookies[0])
		}
	})

	t.Run("user can register and log in", func(t *testing.T) {
		email := "myuser@email.com"
		password := "test123456"

		res, err := http.PostForm(register_url, url.Values{
			"user_name": {"test"},
			"surname":   {"idk"},
			"email":     {email},
			"school_id": {schoolIdStr},
			"password":  {password},
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		got := res.StatusCode
		want := http.StatusCreated
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}

		res, err = http.PostForm(login_url, url.Values{
			"email":    {email},
			"password": {password},
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		gotCode := res.StatusCode
		wantCode := http.StatusOK
		if gotCode != wantCode {
			t.Errorf("Got %d, want %d", gotCode, wantCode)
		}

		gotCookies := res.Cookies()

		gotCookiesLen := len(res.Cookies())
		wantCookiesLen := 1
		if gotCookiesLen != wantCookiesLen {
			t.Errorf("Got %d cookies, wanted %d", gotCookiesLen, wantCookiesLen)
		}

		if gotCookies[0].Name != "token" || !gotCookies[0].HttpOnly || gotCookies[0].SameSite != http.SameSiteStrictMode {
			t.Errorf("Got %s invalid cookie", gotCookies[0])
		}
	})
}
