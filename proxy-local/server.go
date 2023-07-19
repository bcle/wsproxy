package main

import (
	"context"
	"github.com/bcle/wsproxy/pkg/proxyio"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
)

type proxyServer struct {
	subProtocol     string
	remoteUrl       string
	destinationAddr string
}

func (s proxyServer) Handle(conn *net.TCPConn, connId int) {
	defer conn.Close()
	log.Infof("handling new incoming connection %d", connId)
	remote, err := url.Parse(s.remoteUrl)
	if err != nil {
		log.Warnf("failed to parse remote url: %s", err)
		return
	}
	originURL := *remote
	if remote.Scheme == "wss" {
		originURL.Scheme = "https"
	} else {
		originURL.Scheme = "http"
	}
	origin := originURL.String()
	headers := make(http.Header)
	headers.Add("Origin", origin)
	headers.Add("Destination", s.destinationAddr)

	opts := websocket.DialOptions{
		Subprotocols: []string{s.subProtocol},
		HTTPHeader:   headers,
	}
	ws, _, err := websocket.Dial(context.Background(), s.remoteUrl, &opts)
	if err != nil {
		log.Warnf("failed to dial remote proxy: %s", err)
		return
	}
	log.Debugf("[%d] joining connection to web socket", connId)
	proxyio.Join(context.Background(), ws, conn)
	wsClose(ws, websocket.StatusNormalClosure, "websocket closing")
}

func wsClose(ws *websocket.Conn, status websocket.StatusCode, msg string) {
	log.Debug("websocket closing: ", msg)
	ws.Close(status, msg)
}
