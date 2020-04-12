package rtc

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is the Gst Logger Instance
var Logger *logrus.Logger
var rtcpPLIInterval = time.Second * 1
