package main

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	var listenAddr string
	var subProtocol string
	var verbose bool
	var help bool

	flag.StringVar(&listenAddr, "listen-address", ":8088", "The address to listen on")
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
	err := run(subProtocol, listenAddr)
	if err != nil {
		log.Fatal(err)
	}
}

func run(subProtocol string, listenAddr string) error {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	log.Infof("listening on http://%v", l.Addr())

	s := &http.Server{
		Handler:      proxyServer{SubProtocol: subProtocol},
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
}
