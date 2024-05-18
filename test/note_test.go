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

func TestNote(t *testing.T) {
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
	studentId, err := createUser(conn, schoolId)
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
	timetableId, err := createRegularTimetable(conn, periodId, subjectId, roomId, schoolId)
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
