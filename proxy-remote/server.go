package main

import (
	"context"
	"github.com/bcle/wsproxy/pkg/proxyio"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"nhooyr.io/websocket"
)

type proxyServer struct {
	SubProtocol string
}

func (s proxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{s.SubProtocol},
	})
	if err != nil {
		log.Error("failed to accept websocket: %v", err)
		return
	}

	if ws.Subprotocol() != s.SubProtocol {
		wsClose(ws, websocket.StatusPolicyViolation, "client does not speak the expected subprotocol")
		return
	}

	destination := r.Header.Get("destination")
	conn, err := net.Dial("tcp", destination)
	if err != nil {
		wsClose(ws, websocket.StatusBadGateway, "failed to connect to destination")
	}

	tcpConn := conn.(*net.TCPConn)
	proxyio.Join(context.Background(), ws, tcpConn)
	wsClose(ws, websocket.StatusNormalClosure, "websocket closing")
	/*
		err = echo(r.Context(), ws, l)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}
		if err != nil {
			s.logf("failed to echo with %v: %v", r.RemoteAddr, err)
			return
		}
	*/
}

func wsClose(ws *websocket.Conn, status websocket.StatusCode, msg string) {
	log.Debug("websocket closing: ", msg)
	ws.Close(status, msg)
}
