package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type LocalProxyConfig struct {
	RemoteProxyUrl string    `yaml:"remoteProxyUrl"`
	Services       []Service `yaml:"services"`
}

type Service struct {
	Name               string `yaml:"name"`
	LocalAddress       string `yaml:"localAddress"`
	DestinationAddress string `yaml:"destinationAddress"`
}

func LoadFromFile(fname string) (*LocalProxyConfig, error) {
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}
	var cfg LocalProxyConfig
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %s", err)
	}
	return &cfg, nil
}
