package auth

import (
	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/sirupsen/logrus"
	"github.com/wader/gormstore"
)

// Logger is the Auth Logger Instance
var Logger *logrus.Logger

//RegisterWebRoutes for Auth
func RegisterWebRoutes(r *mux.Router, csrfKey []byte, userRepo mgt.Repository, sessionStore *gormstore.Store) *Middleware {
	//Middleware for Auth
	mw := Middleware{SessionStore: sessionStore, CsrfKey: csrfKey, UserRepo: userRepo}
	// SubRouter without Middleware
	router := r.PathPrefix("/").Subrouter()
	router.Use(mw.CsfrMiddleware)
	//Login Routes
	loginController := newWebLoginController(userRepo, sessionStore)
	router.HandleFunc("/login", loginController.GetLogin).Methods("GET")
	router.HandleFunc("/login", loginController.PostLogin).Methods("POST")
	router.HandleFunc("/logout", loginController.GetLogout).Methods("GET")
	return &mw
}

//RegisterAPIRoutes for Auth
func RegisterAPIRoutes(r *mux.Router, sessionManager *rtc.SessionManager, userRepo *mgt.Repository) {

}
