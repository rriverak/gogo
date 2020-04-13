package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/rtc"
)

//RegisterRoutes for the API
func RegisterRoutes(r *mux.Router, sessionManager *rtc.SessionManager) {
	webHandler := WebHandler{StaticDir: "./web/static/", SessionManager: sessionManager}
	webHandler.RegisterWebRoutes(r)
}

type WebHandler struct {
	StaticDir      string
	SessionManager *rtc.SessionManager
}

func (s *WebHandler) RegisterWebRoutes(r *mux.Router) {
	fs := http.FileServer(http.Dir(s.StaticDir))
	r.PathPrefix("/").Handler(fs)
}
