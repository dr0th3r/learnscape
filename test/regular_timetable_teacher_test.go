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

func TestRegularTimetableTeacher(t *testing.T) {
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
	teacherId, err := createUser(conn)
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
	roomId, err := createRoom(conn, teacherId, schoolId)
	if err != nil {
		t.Error(err)
	}
	regularTimetableId, err := createRegularTimetable(conn, periodId, subjectId, schoolId, roomId)
	if err != nil {
		t.Error(err)
	}

	create_url := "http://localhost:8080/regular_timetable_teacher"

	t.Run("can't create regualar_timetable_teacher without regular timetable id", func(t *testing.T) {
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

	t.Run("can't create regular_timetable_teacher without teacher id", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"regular_timetable_id": {regularTimetableId},
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
			"regular_timetable_id": {regularTimetableId},
			"teacher_id":           {teacherId},
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
