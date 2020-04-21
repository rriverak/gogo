package app

import (
	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/pkg/auth"
	"github.com/sirupsen/logrus"
)

// Logger is the API Logger Instance
var Logger *logrus.Logger

//RegisterRoutes for the API
func RegisterRoutes(r *mux.Router, sessionManager *rtc.SessionManager, userRepo *mgt.Repository, mw *auth.Middleware) {
	appHandler := appHandler{SessionManager: sessionManager, UserRepo: userRepo}
	appHandler.RegisterAppRoutes(r, mw)
}

type appHandler struct {
	SessionManager *rtc.SessionManager
	UserRepo       *mgt.Repository
}

func (s *appHandler) RegisterAppRoutes(r *mux.Router, mw *auth.Middleware) {
	router := r.PathPrefix("/").Subrouter()
	router.Use(mw.SessionMiddleware)
	//Start
	startController := NewStartController(s.SessionManager)
	router.HandleFunc("/", startController.Get).Methods("GET")

	//Start
	sessionController := NewSessionController(s.SessionManager)
	router.HandleFunc("/session/create", sessionController.PostNewSession).Methods("POST")
	router.HandleFunc("/session/sdp/{id}", sessionController.PostSDPSession).Methods("POST")
	router.HandleFunc("/session/{id}", sessionController.GetSession).Methods("GET")
}
