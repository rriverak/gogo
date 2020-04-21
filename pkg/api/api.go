package api

import (
	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/sirupsen/logrus"
)

// Logger is the API Logger Instance
var Logger *logrus.Logger

//RegisterRoutes for the API
func RegisterRoutes(r *mux.Router, sessionManager *rtc.SessionManager, usersRepo *mgt.Repository) {
	usersRouter := r.PathPrefix("/api/users").Subrouter()
	sessionRouter := r.PathPrefix("/api/sessions").Subrouter()
	//User
	usersHandler := UsersHandler{UserRepo: *usersRepo}
	usersHandler.RegisterUsersRoutes(usersRouter)
	//Sessions
	sessionHandler := SessionHandler{SessionManager: sessionManager}
	sessionHandler.RegisterSessionRoutes(sessionRouter)

}
