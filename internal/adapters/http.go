package adapters

import (
	"fmt"
	"log"
	"net/http"
	"rebel-shell/internal/app"
)

type HTTPServer struct {
	app *app.RebelshellApp
}

func NewHTTPServer(app *app.RebelshellApp) *HTTPServer {
	return &HTTPServer{app: app}
}

func (s *HTTPServer) Start() {
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		// Example HTTP handler for configuration
		if r.Method == http.MethodGet {
			key := r.URL.Query().Get("key")
			value := s.app.GetConfig(key)
			w.Write([]byte(fmt.Sprintf("%s: %s", key, value)))
		}
	})
	log.Println("Starting HTTP server on :8080")
	http.ListenAndServe(":8080", nil)
}
