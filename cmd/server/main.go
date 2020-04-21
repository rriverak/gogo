package main

import (
	"os"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/gst"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/pkg/api"
	"github.com/rriverak/gogo/pkg/app"
	"github.com/rriverak/gogo/pkg/auth"
	"github.com/rriverak/gogo/pkg/webserver"
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
	mgt.Logger = Logger
	app.Logger = Logger
	auth.Logger = Logger
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

	// StartWebServer
	webserver.Start(cfg)

	// GStreamer MainLoop in MainThread
	gst.StartMainLoop()
}
