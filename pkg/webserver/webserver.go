package webserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/pkg/admin"
	"github.com/rriverak/gogo/pkg/api"
	"github.com/rriverak/gogo/pkg/app"
	"github.com/rriverak/gogo/pkg/auth"
	"github.com/sirupsen/logrus"
	"github.com/wader/gormstore"
)

// Logger is the Server Logger Instance
var Logger *logrus.Logger

//Start a Server
func Start(cfg *config.Config) {
	// User Repository
	db, err := gorm.Open(cfg.DataBase.Driver, cfg.DataBase.ConncetionString)
	if err != nil {
		log.Fatalln(err)
	}
	userRepo := mgt.NewUserRepository(cfg, db)

	// Session Manager
	sessionMgr := rtc.NewSessionManager(cfg)

	// Router
	router := mux.NewRouter()

	// API Routes
	if cfg.Web.API {
		Logger.Info("Web API: Enabled")
		api.RegisterRoutes(router, &sessionMgr, &userRepo)
	} else {
		Logger.Info("Web API: Disabeld")
	}

	// Web Routes
	if cfg.Web.App {
		Logger.Info("Web Interface: Enabled")
		// Sessions
		sessionKey, err := cfg.Web.GetSessionKey()
		if err != nil {
			Logger.Panicf("SessionKey Error: %v", err)
		}
		sessionStore := gormstore.New(db, sessionKey)

		// CSRF
		csrfKey, err := cfg.Web.GetCsrfKey()
		if err != nil {
			Logger.Panicf("CSRFKey Error: %v", err)
		}
		// Middleware
		authMiddleware := auth.RegisterWebRoutes(router, csrfKey, userRepo, sessionStore)

		// Users
		app.RegisterRoutes(router, &sessionMgr, &userRepo, authMiddleware)
		// Admin
		admin.RegisterRoutes(router, &sessionMgr, &userRepo, authMiddleware)

		// Static files
		fs := http.FileServer(http.Dir("./web/static/"))
		router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", fs))

	} else {
		Logger.Info("Web Interface: Disabeld")
	}

	go func() {
		// Start WebServer
		Logger.Infof("HTTP Server bind on '%s' ", cfg.Web.HTTPBind)
		Logger.Infof("Press [ctrl+c] to close the server...")
		err := http.ListenAndServe(cfg.Web.HTTPBind, router)
		Logger.Errorf("HTTP Error: %v", err)
	}()
}
