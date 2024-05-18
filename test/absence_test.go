package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	i "github.com/dr0th3r/learnscape/internal"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5"
)

func TestAbsence(t *testing.T) {
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
	userId, err := createUser(conn, -1)
	if err != nil {
		t.Error(err)
	}

	start := time.Now().Format(time.RFC3339)
	end := time.Now().Add(1 * time.Hour * 168).Format(time.RFC3339)

	create_url := "http://localhost:8080/absence"

	t.Run("can't create absence without user id", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"start": {start},
			"end":   {end},
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

	t.Run("can't create absence without end", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"user_id": {userId},
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

	t.Run("can create absence", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"user_id": {userId},
			"start":   {start},
			"end":     {end},
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
