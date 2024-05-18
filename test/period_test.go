package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"

	i "github.com/dr0th3r/learnscape/internal"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
)

func TestPeriod(t *testing.T) {
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

	//create school for purpose of testing
	conn, err := pgx.Connect(context.Background(), config.DB.GetConnectionUrl())
	if err != nil {
		t.Error(err)
	}
	schoolId, err := createSchool(conn)
	if err != nil {
		t.Error(err)
	}
	userId, err := createUser(conn, schoolId)
	if err != nil {
		t.Error(err)
	}
	claims, err := createUserJWT(userId, schoolId)
	if err != nil {
		t.Error(err)
	}

	create_period_url := "http://localhost:8080/period"

	t.Run("can't create period with invalid time format", func(t *testing.T) {
		formData := url.Values{
			"start": {"08:00:00"},
			"end":   {"08:00:00"},
		}

		formDataReader := strings.NewReader(formData.Encode())

		req, err := http.NewRequest("POST", create_period_url, formDataReader)
		if err != nil {
			t.Error(err)
		}
		req.AddCookie(&claims)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
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

	t.Run("can't create period if end is before start", func(t *testing.T) {
		formData := url.Values{
			"start": {"08:00"},
			"end":   {"07:45"},
		}

		formDataReader := strings.NewReader(formData.Encode())

		req, err := http.NewRequest("POST", create_period_url, formDataReader)
		if err != nil {
			t.Error(err)
		}
		req.AddCookie(&claims)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
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

	t.Run("can create valid period", func(t *testing.T) {
		formData := url.Values{
			"start": {"08:00"},
			"end":   {"08:45"},
		}

		formDataReader := strings.NewReader(formData.Encode())

		req, err := http.NewRequest("POST", create_period_url, formDataReader)
		if err != nil {
			t.Error(err)
		}
		req.AddCookie(&claims)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Error(err)
		}

		got := res.StatusCode
		want := http.StatusCreated
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
			t.Error(string(b))
		}
	})

}
