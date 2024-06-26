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

func TestEventTimetable(t *testing.T) {
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
	name := "test"
	description := "some description"
	start := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	end := time.Now().Add(time.Hour * 24).Format(time.RFC3339)

	schoolIdStr := fmt.Sprint(schoolId)

	create_url := "http://localhost:8080/event_timetable"

	t.Run("incomplete body returns 400 bad request", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"name": {name},
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

	t.Run("can create without description", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"name":      {name},
			"start":     {start},
			"end":       {end},
			"school_id": {schoolIdStr},
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

	t.Run("can create valid event timetable", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"name":        {name},
			"description": {description},
			"start":       {start},
			"end":         {end},
			"school_id":   {schoolIdStr},
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
