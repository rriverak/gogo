package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

//Config Struct
type Config struct {
	Web *WebConfig
}

//WebConfig Struct
type WebConfig struct {
	HTTPBind string
	API      bool
}

//GetDefaultWebConfig Func
func GetDefaultWebConfig() Config {
	return Config{
		Web: &WebConfig{
			HTTPBind: ":8080",
			API:      true,
		},
	}
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
