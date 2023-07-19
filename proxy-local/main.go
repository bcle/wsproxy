package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
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

	flag.StringVar(&listenAddr, "listen-address", ":8087", "The address to listen on")
	flag.StringVar(&remoteUrl, "remote-url", "ws://localhost:8088", "The URL of the remote proxy")
	flag.StringVar(&destinationAddr, "destination-address", "", "The address of the final destination")
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
	if destinationAddr == "" {
		log.Fatal("no destination address specified")
	}
	err := run(subProtocol, listenAddr, remoteUrl, destinationAddr)
	if err != nil {
		log.Fatal(err)
	}
}

func run(subProtocol string, listenAddr string, remoteUrl string, destAddr string) error {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	log.Infof("listening on %v", l.Addr())
	errc := make(chan error, 1)

	srv := &proxyServer{
		subProtocol:     subProtocol,
		remoteUrl:       remoteUrl,
		destinationAddr: destAddr,
	}
	connId := 1
	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accept failed: %s", err)
		}
		tcpConn := conn.(*net.TCPConn)
		go srv.Handle(tcpConn, connId)
		connId += 1
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	/*
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
	*/
	return nil
}
