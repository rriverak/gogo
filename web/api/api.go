package api

import (
	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/sirupsen/logrus"
)

// Logger is the API Logger Instance
var Logger *logrus.Logger

//RegisterRoutes for the API
func RegisterRoutes(r *mux.Router, sessionManager *rtc.SessionManager) {
	router := r.PathPrefix("/api/sessions").Subrouter()

	sessionHandler := SessionHandler{SessionManager: sessionManager}
	sessionHandler.RegisterSessionRoutes(router)

}
