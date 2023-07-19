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
	doneChan chan<- string, // to signal that the source is done
) {
	defer func() { doneChan <- "websocket" }()
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

/*
func copy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 4096)
	for {
		numRead, err2 := src.Read(buf)
		if numRead > 0 {
			log.Debugf("read %d bytes", numRead)
			numWritten, err3 := dst.Write(buf[:numRead])
			if err3 != nil {
				return written, fmt.Errorf("failed to write: %s", err3)
			}
			written += int64(numWritten)
			log.Debugf("wrote %d bytes (%d total)", numWritten, written)
		}
		if err2 != nil {
			log.Debugf("read error: %s", err2)
			return written, err2
		}
	}
	return
}
*/

func copyToWsFromNet(
	ctx context.Context,
	dst *websocket.Conn,
	src *net.TCPConn,
	doneChan chan<- string, // to signal that the source is done
) {
	defer func() { doneChan <- src.RemoteAddr().String() }()

	var err error
	var numWritten int
	buf := make([]byte, 8192)
	for {
		numRead, readErr := src.Read(buf)
		if numRead > 0 {
			log.Debugf("read %d bytes", numRead)
			writeErr := dst.Write(ctx, websocket.MessageBinary, buf[:numRead])
			if writeErr != nil {
				log.Warnf("failed to write to websocket: %s", writeErr)
			} else {
				numWritten += numRead
				log.Debugf("wrote %d bytes (%d total)", numRead, numWritten)
			}
		}
		if readErr != nil {
			err = readErr
			break
		}
	}
	/*
		writer, err := dst.Writer(ctx, websocket.MessageBinary)
		if err != nil {
			log.Errorf("failed to get writer from ws: %s", err)
			return
		}
		defer writer.Close()
		//numWritten, err := io.Copy(writer, src)
		numWritten, err := copy(writer, src)
	*/
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
	doneChan := make(chan string, 2)
	go copyToWsFromNet(ctx, ws, n, doneChan)
	go copyToNetFromWs(ctx, n, ws, doneChan)
	closedSrc := <-doneChan
	log.Infof("1st source to close: %s", closedSrc)
	timer := time.NewTimer(maxTeardownTimeInSeconds * time.Second)
	select {
	case closedSrc = <-doneChan:
		log.Infof("2nd source to close: %s (all done)", closedSrc)
		timer.Stop()
	case <-timer.C:
		log.Warn("timed out waiting for 2nd source to close")
	}
	return
}
