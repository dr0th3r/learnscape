package homepage

import (
	"html/template"
	"net/http"
)

func HandleGet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
