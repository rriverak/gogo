package rtc

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Logger is the Gst Logger Instance
var Logger *logrus.Logger
var rtcpPLIInterval = time.Second * 1

//NewSessionManager create a new SessionManager
func NewSessionManager() SessionManager {
	mgr := SessionManager{}
	mgr.SessionRegister = cache.New(cache.NoExpiration, cache.NoExpiration)
	return mgr
}
