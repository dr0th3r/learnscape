package internal

import (
	"net/http"

	"github.com/dr0th3r/learnscape/internal/class"
	"github.com/dr0th3r/learnscape/internal/group"
	hcheck "github.com/dr0th3r/learnscape/internal/healthCheck"
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
	mux.Handle("POST /regular_report", utils.ParseForm(
		report.HandleCreateRegularReport(db), report.ParseRegularReport,
	))
	mux.Handle("POST /class", utils.ParseForm(
		class.HandleCreateClass(db), class.Parse,
	))
	mux.Handle("POST /group", utils.ParseForm(
		group.HandleCreateGroup(db), group.Parse,
	))
}
