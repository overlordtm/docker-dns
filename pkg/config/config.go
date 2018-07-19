package config

import (
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	TLD            string
	TTL            uint32
	Listen         string
	DockerEndpoint string
	Aliases        map[string]string
}

func ReadConfig(path string) (conf *Config, err error) {
	conf = new(Config)

	c, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(c, conf)
	if err != nil {
		return
	}

	return
}