package proxylocal

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

func Run(
	svcName string,
	subProtocol string,
	listenAddr string,
	remoteUrl string,
	destAddr string,
) {
	log.Infof("starting service <%s>", svcName)
	err := run(svcName, subProtocol, listenAddr, remoteUrl, destAddr)
	if err != nil {
		log.Warnf("service <%s> failed to run: %s", svcName, err)
	}
}

func run(
	svcName string,
	subProtocol string,
	listenAddr string,
	remoteUrl string,
	destAddr string,
) error {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	log.Infof("service <%s> listening on %v", svcName, l.Addr())

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
	return nil
}
