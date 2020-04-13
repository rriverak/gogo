package webserver

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/web/api"
	"github.com/rriverak/gogo/web/app"
	"github.com/sirupsen/logrus"
)

// Logger is the Server Logger Instance
var Logger *logrus.Logger

//Start a Server
func Start(sessionMgr *rtc.SessionManager, cfg *config.Config) {
	// Router
	router := mux.NewRouter()

	// API Routes
	if cfg.Web.API {
		Logger.Info("Web API: Enabled")
		api.RegisterRoutes(router, sessionMgr)
	} else {
		Logger.Info("Web API: Disabeld")
	}

	// Web Routes
	if cfg.Web.App {
		Logger.Info("Web Interface: Enabled")
		app.RegisterRoutes(router, sessionMgr)
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
