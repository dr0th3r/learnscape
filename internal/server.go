package internal

import (
	"net/http"

	"github.com/dr0th3r/learnscape/internal/absence"
	"github.com/dr0th3r/learnscape/internal/class"
	"github.com/dr0th3r/learnscape/internal/grade"
	"github.com/dr0th3r/learnscape/internal/group"
	hcheck "github.com/dr0th3r/learnscape/internal/healthCheck"
	"github.com/dr0th3r/learnscape/internal/note"
	"github.com/dr0th3r/learnscape/internal/period"
	"github.com/dr0th3r/learnscape/internal/report"
	"github.com/dr0th3r/learnscape/internal/room"
	"github.com/dr0th3r/learnscape/internal/school"
	"github.com/dr0th3r/learnscape/internal/subject"
	"github.com/dr0th3r/learnscape/internal/timetable"
	"github.com/dr0th3r/learnscape/internal/user"
	"github.com/dr0th3r/learnscape/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewServer(db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()
	css := http.FileServer(http.Dir("./web/css/"))
	js := http.FileServer(http.Dir("./web/js/"))
	mux.Handle("/css/", http.StripPrefix("/css/", css))
	mux.Handle("/js/", http.StripPrefix("/js/", js))

	addRoutes(mux, db)
	var handler http.Handler = mux
	handler = otelhttp.NewHandler(handler, "server")
	return handler
}

func addRoutes(
	mux *http.ServeMux,
	db *pgxpool.Pool,
) {

	mux.Handle("/health_check", hcheck.HandleHealthCheck())
	mux.Handle("POST /register_user", utils.ParseForm(
		user.HandleRegisterUser(db), user.ParseRegister,
	))
	mux.Handle("POST /login", utils.ParseForm(
		user.HandleLogin(db), user.ParseLogin,
	))
	mux.Handle("POST /register_school", utils.ParseForm(
		school.HandleRegisterSchool(db), user.ParseRegister, school.Parse,
	))
	mux.Handle("POST /period", utils.ParseForm(
		period.HandleCreatePeriod(db), period.Parse,
	))
	mux.Handle("POST /room", utils.ParseForm(
		room.HandleCreateRoom(db), room.Parse,
	))
	mux.Handle("POST /subject", utils.ParseForm(
		subject.HandleCreateSubject(db), subject.Parse,
	))
	mux.Handle("POST /regular_timetable", utils.ParseForm(
		timetable.HandleCreateRegularTimetable(db), timetable.ParseRegularTimetable,
	))
	mux.Handle("POST /substitute_timetable", utils.ParseForm(
		timetable.HandleCreateSubstituteTimetable(db), timetable.ParseSubstituteTimetable,
	))
	mux.Handle("POST /event_timetable", utils.ParseForm(
		timetable.HandleCreateEventTimetable(db), timetable.ParseEventTimetable,
	))
	mux.Handle("POST /report", utils.ParseForm(
		report.HandleCreateReport(db), report.ParseReport,
	))
	mux.Handle("POST /class", utils.ParseForm(
		class.HandleCreateClass(db), class.Parse,
	))
	mux.Handle("POST /group", utils.ParseForm(
		group.HandleCreateGroup(db), group.ParseGroup,
	))
	mux.Handle("POST /users_group", utils.ParseForm(
		group.HandleCreateUsersGroup(db), group.ParseUsersGroup,
	))
	mux.Handle("POST /timetable_group", utils.ParseForm(
		group.HandleCreateTimetableGroup(db),
		group.ParseTimetableGroup,
	))
	mux.Handle("POST /timetable_teacher", utils.ParseForm(
		user.HandleCreateTimetableTeacher(db),
		user.ParseTimetableTeacher,
	))
	mux.Handle("POST /grade", utils.ParseForm(
		grade.HandleCreateGrade(db),
		grade.Parse,
	))
	mux.Handle("POST /note", utils.ParseForm(
		note.HandleCreateNote(db),
		note.Parse,
	))
	mux.Handle("POST /parent_child", utils.ParseForm(
		user.HandleCreateParentChild(db),
		user.ParseParentChild,
	))
	mux.Handle("POST /absence", utils.ParseForm(
		absence.HandleCreateAbsence(db),
		absence.Parse,
	))
	mux.Handle("GET /register", school.HandleGet())
}
