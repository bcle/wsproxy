package proxyio

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"nhooyr.io/websocket"
	"time"
)

const maxTeardownTimeInSeconds = 5

func copyToNetFromWs(
	ctx context.Context,
	dst *net.TCPConn,
	src *websocket.Conn,
	cancel chan<- string, // to signal that the source is done
) {
	defer func() { cancel <- "websocket" }()
	defer dst.CloseWrite()
	var totalWritten int64
	for {
		typ, reader, err := src.Reader(ctx)
		if err != nil {
			log.Debugf("failed to get reader from ws: %s", err)
			break
		}
		if typ != websocket.MessageBinary {
			log.Errorf("unexpected ws message type: %v", typ)
			return
		}
		numWritten, err := io.Copy(dst, reader)
		reason := "EOF"
		if err != nil {
			reason = err.Error()
		}
		totalWritten += numWritten
		log.Debugf("copied %d bytes from ws message to endpoint %s, finished because: %s",
			numWritten, dst.RemoteAddr().String(), reason)
	}
	log.Debugf("copied a total of %d bytes from websocket to endpoint %s",
		totalWritten, dst.RemoteAddr().String())
	return
}

func copyToWsFromNet(
	ctx context.Context,
	dst *websocket.Conn,
	src *net.TCPConn,
	cancel chan<- string, // to signal that the source is done
) {
	defer func() { cancel <- src.RemoteAddr().String() }()
	writer, err := dst.Writer(ctx, websocket.MessageBinary)
	if err != nil {
		log.Errorf("failed to get writer from ws: %s", err)
		return
	}
	defer writer.Close()
	numWritten, err := io.Copy(writer, src)
	reason := "EOF"
	if err != nil {
		reason = err.Error()
	}
	log.Debugf("copied a total of %d bytes from endpoint %s to websocket, finished because: %s",
		numWritten, src.RemoteAddr().String(), reason)
	return
}

// Join a websocket and a network connection, ferry traffic between
// them until both close (or an error occurs).
func Join(ctx context.Context, ws *websocket.Conn, n *net.TCPConn) {
	log.Debugf("joining websocket to network connection for %s",
		n.RemoteAddr().String())
	cancel := make(chan string, 2)
	go copyToWsFromNet(ctx, ws, n, cancel)
	go copyToNetFromWs(ctx, n, ws, cancel)
	closedSrc := <-cancel
	log.Infof("1st source to close: %s", closedSrc)
	timer := time.NewTimer(maxTeardownTimeInSeconds * time.Second)
	select {
	case closedSrc = <-cancel:
		log.Infof("2nd source to close: %s (all done)", closedSrc)
		timer.Stop()
	case <-timer.C:
		log.Warn("timed out waiting for 2nd source to close")
	}
	return
}
