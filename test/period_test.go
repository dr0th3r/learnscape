package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"testing"

	i "github.com/dr0th3r/learnscape/internal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestPeriod(t *testing.T) {
	db_url := os.Getenv("DATABASE_URL")
	db_name := "test_" + fmt.Sprint(rand.Int())
	t.Setenv("DATABASE_NAME", db_name)

	if err := createNewDB(db_url, db_name); err != nil {
		t.Error(err)
		return
	}

	t.Cleanup(func() {
		if err := dropDB(db_url, db_name); err != nil {
			fmt.Println(err)
		}
	})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	go i.Run(ctx)

	if err := waitForReady(ctx); err != nil {
		t.Error(err)
	}

	//create school for purpose of testing
	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("%s%s", db_url, db_name))
	if err != nil {
		t.Error(err)
	}
	id, err := createSchool(conn)
	if err != nil {
		t.Error(err)
	}

	create_period_url := "http://localhost:8080/period"

	t.Run("can't create period withou school_id", func(t *testing.T) {
		res, err := http.PostForm(create_period_url, url.Values{
			"start": {"08:00:00"},
			"end":   {"08:45:00"},
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

	t.Run("can't create period with invalid time format", func(t *testing.T) {
		res, err := http.PostForm(create_period_url, url.Values{
			"school_id": {id},
			"start":     {"08:00"},
			"end":       {"08:00"},
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

	t.Run("can't create period if end is before start", func(t *testing.T) {
		res, err := http.PostForm(create_period_url, url.Values{
			"school_id": {id},
			"start":     {"08:00:00"},
			"end":       {"07:45:00"},
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

	t.Run("can create valid period", func(t *testing.T) {
		res, err := http.PostForm(create_period_url, url.Values{
			"school_id": {id},
			"start":     {"08:00:00"},
			"end":       {"08:45:00"},
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
	})

}

func createSchool(db *pgx.Conn) (string, error) {
	id := uuid.NewString()
	_, err := db.Exec(context.Background(), "insert into school (id, name, city, zip_code, street_address) values ($1, $2, $3, $4, $5)",
		id, "test", "test city", "123 45", "street 7",
	)
	if err != nil {
		return "", err
	}
	return id, nil
}
