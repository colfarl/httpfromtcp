// Package server used to accept requests and send responses
package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/colfarl/httpfromtcp/internal/request"
	"github.com/colfarl/httpfromtcp/internal/response"
)
type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode	response.StatusCode
	Message		string
}
type Server struct {
	Available	*atomic.Bool
	Listener	net.Listener
}

func RespondWithError(w io.Writer, herr HandlerError) {
    // Minimal plaintext body
    body := []byte(herr.Message)

    // Write a proper status line and headers. These MUST include \r\n.
    _ = response.WriteStatusLine(w, herr.StatusCode)
    _ = response.WriteHeaders(w, response.GetDefaultHeaders(len(body)))

    // Body
    _, _ = w.Write(body)
}

func Serve(port int, handle Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}
	
	
	server :=  &Server{
		Listener: l,
		Available: &atomic.Bool{},
	}

	server.Available.Store(true)
	go server.listen(handle)

	return server, nil
}

func (s *Server) Close() error {
	if !s.Available.CompareAndSwap(true, false){
		return nil
	}
	if s.Listener == nil {
		return fmt.Errorf("listener does not exist")
	}
	s.Listener.Close()
	return nil
}

func (s *Server) listen(handler Handler) {
	for {
		if !s.Available.Load() {
			continue
		}

		conn, err := s.Listener.Accept()
		if err != nil {
			if !s.Available.Load(){
				return
			}
			log.Print("uh oh:", err, "\n")
			continue
		}
		go s.handle(conn, handler)
	}
}

func (s *Server) handle(conn net.Conn, handler Handler) {
	r, err := request.RequestFromReader(conn)

	if err != nil {
		rerr := HandlerError{
			StatusCode: response.BadRequest,
			Message: "improperly formatted request",
		}
		RespondWithError(conn, rerr)
		conn.Close()
		return
	}
	
	buffer := bytes.NewBuffer([]byte{})
	herr := handler(buffer, r)
	if herr != nil {
		RespondWithError(conn, *herr)
		conn.Close()
		return
	}
	
	body := buffer.String()	
	headers := response.GetDefaultHeaders(len(body))
	response.WriteStatusLine(conn, response.OK)
	response.WriteHeaders(conn, headers)
	conn.Write([]byte(body))	
	conn.Close()
}
