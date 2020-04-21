package admin

import (
	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/utils"
	"github.com/rriverak/gogo/pkg/auth"
)

//RegisterRoutes for Admin Package
func RegisterRoutes(r *mux.Router, sessionManager *rtc.SessionManager, userRepo *mgt.Repository, mw *auth.Middleware) {
	router := r.PathPrefix("/admin/").Subrouter()
	//Set Session Middleware
	router.Use(mw.SessionMiddleware)
	//Nav
	navBuilder := utils.NewNavBuilder()
	//Dashboard
	dashboardController := NewDashboardController(sessionManager, navBuilder)
	router.HandleFunc("/", dashboardController.Get).Methods("GET")
	navBuilder.AddElement("Dashboard", "/admin/", "fa-chart-area")
	//Sessions
	sessionsController := NewSessionsController(sessionManager, navBuilder)
	router.HandleFunc("/sessions", sessionsController.Get).Methods("GET")
	navBuilder.AddElement("Sessions", "/admin/sessions", "fa-video")
	//Users
	usersController := NewUsersController(sessionManager, navBuilder, *userRepo)
	router.HandleFunc("/users", usersController.GetList).Methods("GET")          // LIST PAGE
	router.HandleFunc("/users/create", usersController.GetUser).Methods("GET")   // CREATE PAGE
	router.HandleFunc("/users/create", usersController.PostUser).Methods("POST") // SAVE CREATE
	router.HandleFunc("/users/{id}", usersController.GetUser).Methods("GET")     // EDIT PAGE
	router.HandleFunc("/users/{id}", usersController.PostUser).Methods("POST")   // SAVE EDIT

	navBuilder.AddElement("Users", "/admin/users", "fa-users")
}
