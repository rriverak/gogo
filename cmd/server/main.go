package main

import (
	"os"

	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/gst"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/web/api"
	"github.com/rriverak/gogo/web/webserver"
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
	webserver.Logger = Logger
}

func main() {
	// Name & Version
	Logger.Info("Video GroupCall Server v0.1")
	// Config
	cfg := config.LoadOrCreateConfig("config.yaml")

	// Logging
	Logger.Infof("LogLevel: %v", cfg.LogLevel)
	Logger.SetLevel(cfg.GetLogLevel())

	// Session Manager
	sessionMgr := rtc.NewSessionManager(cfg)

	// StartWebServer
	webserver.Start(&sessionMgr, cfg)

	// GStreamer MainLoop in MainThread
	gst.StartMainLoop()
}
