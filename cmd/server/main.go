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

func main() {
	// Config
	cfg := config.LoadConfig("config.yaml")
	if cfg == nil {
		dCfg := config.GetDefaultWebConfig()
		config.SaveConfig("config.yaml", dCfg)
		cfg = &dCfg
	}
	// Logger
	var logger = logrus.New()
	logger.Out = os.Stdout
	api.Logger = logger
	gst.Logger = logger
	rtc.Logger = logger

	// Router
	router := mux.NewRouter()

	// API
	sessionHandler := api.SessionHandler{}
	sessionHandler.RegisterSessionRoutes(router)

	// WEB
	webHandler := web.WebHandler{StaticDir: "./web/static/"}
	webHandler.RegisterWebRoutes(router)
	go func() {
		// WebServer
		logger.Infof("HTTP Listen on: %s \n", cfg.Web.HTTPBind)
		logger.Infof("Press [ctrl+c] to close the server...")
		logger.Infof("Binding Error: %v \n", http.ListenAndServe(cfg.Web.HTTPBind, router))

	}()

	gst.StartMainLoop()
}
