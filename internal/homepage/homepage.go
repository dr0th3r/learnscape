package homepage

import (
	"html/template"
	"net/http"
)

func HandleGet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./web/homepage.html"))
		tmpl.Execute(w, nil)
	})
}
