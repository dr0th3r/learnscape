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

func TestTimetableGroup(t *testing.T) {
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
	userId, err := createUser(conn, schoolId)
	if err != nil {
		t.Error(err)
	}
	groupId, err := createGroup(conn)
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
	roomId, err := createRoom(conn, userId, schoolId)
	if err != nil {
		t.Error(err)
	}
	timetableId, err := createRegularTimetable(conn, periodId, subjectId, roomId, schoolId)
	if err != nil {
		t.Error(err)
	}

	create_url := "http://localhost:8080/timetable_group"

	t.Run("can't create timetable_group without  timetable id", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"group_id": {groupId},
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

	t.Run("can't create timetable_group without group id", func(t *testing.T) {
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

	t.Run("can create timetable_group", func(t *testing.T) {
		res, err := http.PostForm(create_url, url.Values{
			"timetable_id": {timetableId},
			"group_id":     {groupId},
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
