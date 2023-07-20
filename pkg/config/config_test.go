package config

import (
	"gopkg.in/yaml.v3"
	"testing"
)

func TestMarshal(t *testing.T) {
	cfg := LocalProxyConfig{
		RemoteProxyUrl: "ws://some.where.com",
		Services: []Service{
			{
				Name:               "google",
				LocalAddress:       "192.168.2.105:443",
				DestinationAddress: "www.google.com:443",
			},
			{
				Name:               "facebook",
				LocalAddress:       "192.168.2.106:443",
				DestinationAddress: "www.facebook.com:443",
			},
		},
	}
	s, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	t.Log(string(s))
}
