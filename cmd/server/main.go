package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/api"
	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/gst"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/web"
	"github.com/sirupsen/logrus"
)

//Logger for the Main Package
var Logger = logrus.New()

func init() {
	Logger.Out = os.Stdout
	// Set Logger for all other Packages
	api.Logger = Logger
	gst.Logger = Logger
	rtc.Logger = Logger
}
func main() {
	// Name & Version
	Logger.Info("GOGO Video GroupCall Server - 0.1")
	// Config
	cfg := config.LoadOrCreateConfig("config.yaml")

	// Session Manager
	sessionMgr := rtc.SessionManager{Config: cfg}

	// Router
	router := mux.NewRouter()

	// API Routes
	if cfg.Web.API {
		Logger.Info("Web API: Enabled")
		api.RegisterRoutes(router, &sessionMgr)
	} else {
		Logger.Info("Web API: Disabeld")
	}

	// Web Routes
	if cfg.Web.App {
		Logger.Info("Web Interface: Enabled")
		web.RegisterRoutes(router, &sessionMgr)
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

	// GStreamer MainLoop in MainThread
	gst.StartMainLoop()
}
