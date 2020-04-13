package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/rtc"
)

//RegisterRoutes for the API
func RegisterRoutes(r *mux.Router, sessionManager *rtc.SessionManager) {
	appHandler := appHandler{StaticDir: "./web/static/", SessionManager: sessionManager}
	appHandler.RegisterAppRoutes(r)
}

type appHandler struct {
	StaticDir      string
	SessionManager *rtc.SessionManager
}

func (s *appHandler) RegisterAppRoutes(r *mux.Router) {
	fs := http.FileServer(http.Dir(s.StaticDir))
	r.PathPrefix("/").Handler(fs)
}
