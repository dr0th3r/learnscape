package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetHomepage(db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, span := tracer.Start(reqCtx, "get homepage")
		defer span.End()

		claims := reqCtx.Value("claims").(utils.UserClaims)
		peridRows, err := db.Query(ctx, "SELECT start, end FROM period WHERE school_id=$1", claims.SchoolId)
		if err != nil {
			utils.UnexpectedError(w, err, ctx)
			return
		}
		defer peridRows.Close()

		for peridRows.Next() {
			var start, end string
			if err := peridRows.Scan(&start, &end); err != nil {
				utils.UnexpectedError(w, err, ctx)
				return
			}
			fmt.Println("Start: ", start, "End: ", end)
		}

		randomData := []string{"12:30 - 13:45", "12:30 - 13:45"}
		timetable := struct {
			Periods   []string
			Monday    []string
			Tuesday   []string
			Wednesday []string
			Thursday  []string
			Friday    []string
		}{
			Periods:   randomData,
			Monday:    randomData,
			Tuesday:   randomData,
			Wednesday: randomData,
			Thursday:  randomData,
			Friday:    randomData,
		}
		// timetable["periods"] = []string{"12:30 - 13:45", "12:30 - 13:45"}
		// timetable["monday"] = []string{"12:30 - 13:45", "12:30 - 13:45"}
		// timetable["tuesday"] = []string{"12:30 - 13:45", "12:30 - 13:45"}
		// timetable["wednesday"] = []string{"12:30 - 13:45", "12:30 - 13:45"}
		// timetable["thursday"] = []string{"12:30 - 13:45", "12:30 - 13:45"}
		// timetable["friday"] = []string{"12:30 - 13:45", "12:30 - 13:45"}

		tmpl := template.Must(template.ParseFiles("./web/homepage.html"))
		tmpl.Execute(w, timetable)
	})
}
