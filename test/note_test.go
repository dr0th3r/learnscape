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

func TestNote(t *testing.T) {
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
	studentId, err := createUser(conn)
	if err != nil {
		t.Error(err)
	}
	schoolId, err := createSchool(conn)
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
	roomId, err := createRoom(conn, studentId, schoolId)
	if err != nil {
		t.Error(err)
	}
	timetableId, err := createRegularTimetable(conn, periodId, subjectId, schoolId, roomId)
	if err != nil {
		t.Error(err)
	}

	create_url := "http://localhost:8080/note"

	//date := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	date := "2024-05-08"

	t.Run("can't create note without timetable_id", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"type":    {"homework"},
			"content": {"testing note"},
			"date":    {date},
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

	//NOTE: add more tests later

	t.Run("can create note", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"timetable_id": {timetableId},
			"type":         {"homework"},
			"content":      {"testing note"},
			"date":         {date},
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
