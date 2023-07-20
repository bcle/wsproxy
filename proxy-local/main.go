package main

import (
	"flag"
	"github.com/bcle/wsproxy/pkg/config"
	"github.com/bcle/wsproxy/pkg/proxylocal"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

func main() {
	var listenAddr string
	var subProtocol string
	var verbose bool
	var help bool
	var remoteUrl string
	var destinationAddr string
	var configFile string

	flag.StringVar(&configFile, "config", "", "Use specified config file")
	flag.StringVar(&listenAddr, "listen-address", ":8087", "The host:port address to listen on")
	flag.StringVar(&remoteUrl, "remote-url", "ws://localhost:8088", "The URL of the remote proxy")
	flag.StringVar(&destinationAddr, "destination-address", "", "The host:port address of the final destination")
	flag.StringVar(&subProtocol, "subprotocol", "wsproxy", "The websocket subprotocol name")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging")
	flag.BoolVar(&help, "help", false, "Show usage")
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	var cfg *config.LocalProxyConfig
	var err error
	if configFile != "" {
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			log.Fatalf("failed to load config file: %s", err)
		}
	} else {
		if destinationAddr == "" {
			log.Fatal("no destination address specified")
		}
		cfg = &config.LocalProxyConfig{
			RemoteProxyUrl: remoteUrl,
			Services: []config.Service{
				{
					Name:               "default",
					LocalAddress:       listenAddr,
					DestinationAddress: destinationAddr,
				},
			},
		}
	}
	for _, svc := range cfg.Services {
		go proxylocal.Run(svc.Name, subProtocol, svc.LocalAddress,
			cfg.RemoteProxyUrl, svc.DestinationAddress)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

}
