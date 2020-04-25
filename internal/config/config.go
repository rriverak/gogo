package config

import (
	"encoding/base64"
	"io/ioutil"
	"log"

	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//Config Struct
type Config struct {
	LogLevel string
	DataBase *DataBaseConfig
	Web      *WebConfig
	Media    *MediaConfig
	WebRTC   *WebRtcConfig
}

//GetLogLevel from Config
func (c *Config) GetLogLevel() logrus.Level {
	switch c.LogLevel {
	case "Info":
		return logrus.InfoLevel
	case "Warn":
		return logrus.WarnLevel
	case "Error":
		return logrus.ErrorLevel
	case "Debug":
		return logrus.DebugLevel
	}
	return logrus.InfoLevel
}

//DataBaseConfig Struct
type DataBaseConfig struct {
	Driver           string
	ConncetionString string
}

//WebConfig Struct
type WebConfig struct {
	HTTPBind   string
	API        bool
	App        bool
	SessionKey string
	CsrfKey    string
}

//GetSessionKey from Config
func (w *WebConfig) GetSessionKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(w.SessionKey)
}

//GetCsrfKey from Config
func (w *WebConfig) GetCsrfKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(w.CsrfKey)
}

//MediaConfig Struct
type MediaConfig struct {
	Video *MediaVideoConfig
	Audio *MediaAudioConfig
}

//MediaAudioConfig Struct
type MediaAudioConfig struct {
	Enabled bool
}

//MediaVideoConfig Struct
type MediaVideoConfig struct {
	Enabled     bool
	Codecs      *MediaVideoCodecConfig
	VideoSize   int
	VideoBorder int
}

//GetFullSize get the full Video Size
func (mvc *MediaVideoConfig) GetFullSize() int {
	return mvc.VideoSize + (2 * mvc.VideoBorder)
}

//MediaVideoCodecConfig Struct
type MediaVideoCodecConfig struct {
	VP8  bool
	VP9  bool
	H264 bool
}

//WebRtcConfig Struct
type WebRtcConfig struct {
	ICEServers []string
}

//GetDefaultConfig Func
func GetDefaultConfig() Config {
	return Config{
		LogLevel: "Info",
		DataBase: &DataBaseConfig{
			Driver:           "sqlite3",
			ConncetionString: ":memory:",
		},
		Web: &WebConfig{
			HTTPBind:   ":8080",
			API:        true,
			App:        true,
			SessionKey: base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
			CsrfKey:    base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
		},
		Media: &MediaConfig{
			Video: &MediaVideoConfig{
				Enabled:     true,
				Codecs:      &MediaVideoCodecConfig{VP8: true, H264: true, VP9: false},
				VideoSize:   320,
				VideoBorder: 15,
			},
			Audio: &MediaAudioConfig{
				Enabled: true,
			},
		},
		WebRTC: &WebRtcConfig{
			ICEServers: []string{"stun:stun.l.google.com:19302"},
		},
	}
}

//LoadOrCreateConfig with Default Values
func LoadOrCreateConfig(fileName string) *Config {
	cfg := LoadConfig(fileName)
	if cfg == nil {
		dCfg := GetDefaultConfig()
		SaveConfig(fileName, dCfg)
		cfg = &dCfg
	}
	return cfg
}

//LoadConfig from FileName
func LoadConfig(fileName string) *Config {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil
	}
	cfg := Config{}
	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		return nil
	}
	return &cfg
}

//SaveConfig to FileName
func SaveConfig(fileName string, cfg Config) {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Panic(err)
	}
	ioutil.WriteFile(fileName, data, 0660)
}
