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
	"github.com/jackc/pgx/v5"
)

func TestRegularTimetable(t *testing.T) {
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

	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("%s%s", db_url, db_name))
	if err != nil {
		t.Error(err)
	}
	schoolId, err := createSchool(conn)
	if err != nil {
		t.Error(err)
	}
	teacherId, err := createUser(conn)
	if err != nil {
		t.Error(err)
	}
	periodId, err := createPeriod(conn, schoolId)
	if err != nil {
		t.Error(err)
	}
	subjectId, err := createSubject(conn)
	if err != nil {
		t.Error(err)
	}
	roomId, err := createRoom(conn, teacherId, schoolId)
	if err != nil {
		t.Error(err)
	}
	create_url := "http://localhost:8080/regular_timetable"

	t.Run("incomplete body returns 400 bad request", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"weekday": {"1"},
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

	t.Run("ids must be numbers", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"period_id":  {"1"},
			"subject_id": {"1"},
			"room_id":    {"random"},
			"weekday":    {"1"},
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

	t.Run("invalid weekday returns 400 bad request", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"period_id":  {"1"},
			"subject_id": {"1"},
			"room_id":    {"1"},
			"weekday":    {"random"},
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

	t.Run("can create valid regular timetable", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"period_id":  {periodId},
			"subject_id": {subjectId},
			"room_id":    {roomId},
			"weekday":    {"1"},
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
