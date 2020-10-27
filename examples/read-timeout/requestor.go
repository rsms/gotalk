package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/rsms/gotalk"
)

// slowWriter simulates slow writing, to demonstrate timeout
type slowWriter struct {
	c     io.ReadWriteCloser
	delay time.Duration
}

func (rwc *slowWriter) Read(p []byte) (n int, err error) {
	n, err = rwc.c.Read(p)
	if err == nil {
		// Note: This is a completely unreliable way of detecting messages. It works because
		// we are not sending any payloads starting with the byte 'e'.
		if p[0] == byte(gotalk.MsgTypeRetryRes) {
			fmt.Printf("requestor: Server asked us to retry the request\n")
		}
	}
	return
}

func (rwc *slowWriter) Write(p []byte) (n int, err error) {
	// Delay anything but writing single-request headers

	// tname := "?"
	// switch gotalk.MsgType(p[0]) {
	// case gotalk.MsgTypeSingleReq: tname = "SingleReq"
	// case gotalk.MsgTypeStreamReq: tname = "StreamReq"
	// case gotalk.MsgTypeStreamReqPart: tname = "StreamReqPart"
	// case gotalk.MsgTypeSingleRes: tname = "SingleRes"
	// case gotalk.MsgTypeStreamRes: tname = "StreamRes"
	// case gotalk.MsgTypeErrorRes: tname = "ErrorRes"
	// case gotalk.MsgTypeRetryRes: tname = "RetryRes"
	// case gotalk.MsgTypeNotification: tname = "Notification"
	// case gotalk.MsgTypeHeartbeat: tname = "Heartbeat"
	// case gotalk.MsgTypeProtocolError: tname = "ProtocolError"
	// default:
	// 	fmt.Printf("slowWriter.Write (rest) %q\n", p)
	// 	return rwc.c.Write(p)
	// }
	// fmt.Printf("slowWriter.Write %s\n", tname)

	if err == nil && p[0] == byte(gotalk.MsgTypeSingleReq) {
		time.Sleep(rwc.delay)
		n, err = rwc.c.Write(p)
		return
	}
	return rwc.c.Write(p)
}

func (rwc *slowWriter) Close() error {
	return rwc.c.Close()
}

func sendRequest(s *gotalk.Sock) error {
	fmt.Printf("requestor: sending 'echo' request\n")
	b, err := s.BufferRequest("echo", []byte("Hello"))
	if err == gotalk.ErrTimeout {
		fmt.Printf("requestor: timed out\n")
	} else if err != nil {
		fmt.Printf("requestor: error %v\n", err.Error())
	} else {
		fmt.Printf("requestor: success: %q\n", b)
	}
	return err
}

func sendSlowWritingRequest(port string) error {
	s, err := gotalk.Connect("tcp", "localhost:"+port)
	if err == nil {
		fmt.Printf("requestor: connected to %q on Sock@%p\n", s.Addr(), s)

		// Wrap the connection for slow writing to simulate a poor connection
		s.Adopt(&slowWriter{c: s.Conn(), delay: 200 * time.Millisecond})

		// Send a request -- it will take too long and time out
		fmt.Printf("requestor: send slow request\n")
		err = sendRequest(s)

		s.Close()
	}
	return err
}

func sendRegularRequest(port string) error {
	s, err := gotalk.Connect("tcp", "localhost:"+port)
	if err == nil {
		fmt.Printf("requestor: connected to %q on Sock@%p\n", s.Addr(), s)
		err = sendRequest(s)
		s.Close()
	}
	return err
}

func requestor(port string) {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			err := sendSlowWritingRequest(port)
			if err == nil {
				panic(fmt.Sprintf("expected error from sendSlowWritingRequest (no error)"))
			} else {
				s := err.Error()
				if !strings.Contains(s, "timeout") && !strings.Contains(s, "socket closed") {
					panic(fmt.Sprintf(
						"expected timeout or socket closed from sendSlowWritingRequest but got %v",
						err))
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// time.Sleep(10 * time.Millisecond)
	// if err := sendRegularRequest(port); err != nil {
	// 	panic(fmt.Sprintf("expected sendRegularRequest to succeed but got error %v", err))
	// }
}
