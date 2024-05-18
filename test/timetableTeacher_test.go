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

func TestTimetableTeacher(t *testing.T) {
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

	create_url := "http://localhost:8080/timetable_teacher"

	t.Run("can't create regualar_timetable_teacher without  timetable id", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"teacher_id": {teacherId},
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

	t.Run("can't create timetable_teacher without teacher id", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"timetable_id": {timetableId},
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

	t.Run("can create regualar_timetable_teacher", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"timetable_id": {timetableId},
			"teacher_id":   {teacherId},
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
