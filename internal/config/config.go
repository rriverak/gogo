package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Web *WebConfig
}

type WebConfig struct {
	HTTPBind string
	API      bool
}

func GetDefaultWebConfig() Config {
	return Config{
		Web: &WebConfig{
			HTTPBind: ":82",
			API:      true,
		},
	}
}

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

func SaveConfig(fileName string, cfg Config) {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Panic(err)
	}
	ioutil.WriteFile(fileName, data, 0660)
}
