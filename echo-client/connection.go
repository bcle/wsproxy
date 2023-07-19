package main

import (
	"context"
	"fmt"
	"github.com/chzyer/readline"
	"net/http"
	"nhooyr.io/websocket"
)

type session struct {
	ws      *websocket.Conn
	errChan chan error
}

func connect(url, origin string, rlConf *readline.Config, allowInsecure bool) error {
	headers := make(http.Header)
	headers.Add("Origin", origin)
	headers.Add("Destination", "www.google.com:80")

	opts := websocket.DialOptions{
		Subprotocols: []string{"wsproxy"},
		HTTPHeader:   headers,
	}
	ws, _, err := websocket.Dial(context.Background(), url, &opts)
	if err != nil {
		return err
	}

	/*
		rl, err := readline.NewEx(rlConf)
		if err != nil {
			return err
		}
		defer rl.Close()
	*/
	sess := &session{
		ws:      ws,
		errChan: make(chan error),
	}

	go sess.readConsole()
	go sess.readWebsocket()

	return <-sess.errChan
}

func (s *session) readConsole() {
	request := "GET / HTTP/1.1\r\nHost: www.google.com\r\nUser-Agent: curl/7.76.1\r\nAccept: */*\r\n\r\n"
	err := s.ws.Write(context.Background(), websocket.MessageBinary, []byte(request))
	if err != nil {
		s.errChan <- err
		return
	}

	/*
		for {
			line, err := s.rl.Readline()
			if err != nil {
				s.errChan <- err
				return
			}

			err = s.ws.Write(context.Background(), websocket.MessageBinary, []byte(line))
			if err != nil {
				s.errChan <- err
				return
			}
		}
	*/
}

/*
func bytesToFormattedHex(bytes []byte) string {
	text := hex.EncodeToString(bytes)
	return regexp.MustCompile("(..)").ReplaceAllString(text, "$1 ")
}
*/

func (s *session) readWebsocket() {
	// rxSprintf := color.New(color.FgGreen).SprintfFunc()

	_, reader, err := s.ws.Reader(context.Background())
	if err != nil {
		s.errChan <- err
		return
	}
	var totalBytes int
	for {
		buf := make([]byte, 1024)
		numBytes, err := reader.Read(buf)
		if numBytes > 0 {
			totalBytes += numBytes
			fmt.Println()
			fmt.Printf("read %d bytes (%d total)\n", numBytes, totalBytes)
			fmt.Printf("%s\n\n", buf[:numBytes])
		}
		if err != nil {
			fmt.Println("error:", err)
			break
		}
	}
	/*
		numBytes, err := io.Copy(os.Stdout, reader)
		if err != nil {
			s.errChan <- fmt.Errorf("copy failed: %s", err)
			return
		}
	*/
	fmt.Println("copied a total of", totalBytes, "bytes")
	/*
		for {
			msgType, buf, err := s.ws.Read(context.Background())
			if err != nil {
				s.errChan <- err
				return
			}

			var text string
			switch msgType {
			case websocket.MessageText:
				text = string(buf)
			case websocket.MessageBinary:
				text = bytesToFormattedHex(buf)
			default:
				s.errChan <- fmt.Errorf("unknown websocket frame type: %d", msgType)
				return
			}

			fmt.Fprint(s.rl.Stdout(), rxSprintf("< %s\n", text))
		}
	*/
}
