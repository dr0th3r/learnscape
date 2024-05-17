package internal

import (
	"net/http"

	c "github.com/dr0th3r/learnscape/internal/controllers"
	m "github.com/dr0th3r/learnscape/internal/models"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewServer(db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()
	css := http.FileServer(http.Dir("./web/css/"))
	js := http.FileServer(http.Dir("./web/js/"))
	mux.Handle("GET /css/", http.StripPrefix("/css/", css))
	mux.Handle("GET /js/", http.StripPrefix("/js/", js))

	addRoutes(mux, db)
	var handler http.Handler = mux
	handler = otelhttp.NewHandler(handler, "server")
	return handler
}

func addRoutes(
	mux *http.ServeMux,
	db *pgxpool.Pool,
) {

	mux.Handle("GET /health_check", c.HealthCheck())
	mux.Handle("POST /register_user", utils.ParseForm(
		c.RegisterUser(db), m.ParseRegister,
	))
	mux.Handle("POST /login", utils.ParseForm(
		c.Login(db), m.ParseLogin,
	))
	mux.Handle("POST /register_school", utils.ParseForm(
		c.RegisterSchool(db), m.ParseRegister, m.ParseSchool,
	))
	mux.Handle("POST /period",
		utils.WithAuth(utils.ParseForm(
			c.CreatePeriod(db), m.ParsePeriod,
		)),
	)
	mux.Handle("POST /room", utils.ParseForm(
		c.CreateRoom(db), m.ParseRoom,
	))
	mux.Handle("POST /subject", utils.ParseForm(
		c.CreateSubject(db), m.ParseSubject,
	))
	mux.Handle("POST /regular_timetable", utils.ParseForm(
		c.CreateRegularTimetable(db), m.ParseRegularTimetable,
	))
	mux.Handle("POST /substitute_timetable", utils.ParseForm(
		c.CreateSubstituteTimetable(db), m.ParseSubstituteTimetable,
	))
	mux.Handle("POST /event_timetable", utils.ParseForm(
		c.CreateEventTimetable(db), m.ParseEventTimetable,
	))
	mux.Handle("POST /report", utils.ParseForm(
		c.CreateReport(db), m.ParseReport,
	))
	mux.Handle("POST /class", utils.ParseForm(
		c.CreateClass(db), m.ParseClass,
	))
	mux.Handle("POST /group", utils.ParseForm(
		c.CreateGroup(db), m.ParseGroup,
	))
	mux.Handle("POST /users_group", utils.ParseForm(
		c.CreateUsersGroup(db), m.ParseUsersGroup,
	))
	mux.Handle("POST /timetable_group", utils.ParseForm(
		c.CreateTimetableGroup(db),
		m.ParseTimetableGroup,
	))
	mux.Handle("POST /timetable_teacher", utils.ParseForm(
		c.CreateTimetableTeacher(db),
		m.ParseTimetableTeacher,
	))
	mux.Handle("POST /grade", utils.ParseForm(
		c.CreateGrade(db),
		m.ParseGrade,
	))
	mux.Handle("POST /note", utils.ParseForm(
		c.CreateNote(db), m.ParseNote,
	))
	mux.Handle("POST /parent_child", utils.ParseForm(
		c.CreateParentChild(db), m.ParseParentChild,
	))
	mux.Handle("POST /absence", utils.ParseForm(
		c.CreateAbsence(db), m.ParseAbsence,
	))
	mux.Handle("GET /", utils.WithAuth(c.GetHomepage(db)))
	mux.Handle("GET /register", c.GetRegister())
	mux.Handle("GET /login", c.GetLogin())
}
