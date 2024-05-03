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

func TestRoom(t *testing.T) {
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
	school_id, err := createSchool(conn)
	if err != nil {
		t.Error(err)
	}
	teacher_id, err := createUser(conn)
	if err != nil {
		t.Error(err)
	}

	create_room_url := "http://localhost:8080/room"

	t.Run("can't create room with invalid body", func(t *testing.T) {
		res, err := http.PostForm(create_room_url, url.Values{
			"school_id":  {school_id},
			"teacher_id": {teacher_id},
			//name is missing
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

	t.Run("can create valid room", func(t *testing.T) {
		res, err := http.PostForm(create_room_url, url.Values{
			"school_id":  {school_id},
			"teacher_id": {teacher_id},
			"name":       {"my room"},
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
