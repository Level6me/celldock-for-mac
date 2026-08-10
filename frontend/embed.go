package frontend

import (
	"embed"
	"net/http"
)

//go:embed index.html
var staticFS embed.FS

func RegisterStaticRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			data, err := staticFS.ReadFile("index.html")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write(data)
			return
		}
		http.NotFound(w, r)
	})
}
