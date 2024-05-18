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
)

func TestSchool(t *testing.T) {
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

	register_url := "http://localhost:8080/register_school"

	t.Run("incomplete body returns 400 bad request", func(t *testing.T) {
		res, err := http.PostForm(register_url, url.Values{
			"school_name": {"test"},
			"city":        {"idk"},
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

	t.Run("valid request creates school and admin account", func(t *testing.T) {
		res, err := http.PostForm(register_url, url.Values{
			"school_name":    {"test"},
			"city":           {"idk"},
			"zip_code":       {"123 45"},
			"street_address": {"test 8"},
			//admin info
			"user_name": {"test"},
			"surname":   {"idk"},
			"email":     {"random2@email.com"},
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

		gotCookiesLen := len(gotCookies)
		wantCookiesLen := 1
		if gotCookiesLen != wantCookiesLen {
			t.Errorf("Got %d cookies, wanted %d", gotCookiesLen, wantCookiesLen)
		}

		if gotCookies[0].Name != "token" || !gotCookies[0].HttpOnly || gotCookies[0].SameSite != http.SameSiteStrictMode {
			t.Errorf("Got %s invalid cookie", gotCookies[0])
		}
	})
}
