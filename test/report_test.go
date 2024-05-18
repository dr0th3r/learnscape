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

func TestRegularReport(t *testing.T) {
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
	teacherId, err := createUser(conn, schoolId)
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

	timetableId, err := createRegularTimetable(conn, periodId, subjectId, roomId, schoolId)
	if err != nil {
		t.Error(err)
	}

	create_url := "http://localhost:8080/report"

	t.Run("incomplete body returns 400 bad request", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"topic_covered": {"linear algebra"},
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

	t.Run("regular_timetable_id must be numbers", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"regular_timetable_id": {"idk"},
			"topic_covered":        {"linear algebra"},
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

	t.Run("can create valid regular report", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"timetable_id":  {timetableId},
			"reported_by":   {teacherId},
			"topic_covered": {"linear algebra"},
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
