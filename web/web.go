package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

type WebHandler struct {
	StaticDir string
}

func (s *WebHandler) RegisterWebRoutes(r *mux.Router) {
	fs := http.FileServer(http.Dir(s.StaticDir))
	r.PathPrefix("/").Handler(fs)
}
