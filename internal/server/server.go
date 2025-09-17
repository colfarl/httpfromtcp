// Package server used to accept requests and send responses
package server

import (
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

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		return nil, err
	}
	
	
	server :=  &Server{
		Listener: l,
		Available: &atomic.Bool{},
	}

	server.Available.Store(true)
	go server.listen()

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

func (s *Server) listen() {
	for {
		if !s.Available.Load() {
			continue
		}

		conn, err := s.Listener.Accept()
		if err != nil {
			log.Print("uh oh:", err, "\n")
			if !s.Available.Load(){
				return
			}
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	s.Available.Store(false)	
	response.WriteStatusLine(conn, response.OK)
	headers := response.GetDefaultHeaders(0)
	response.WriteHeaders(conn, headers)
	conn.Close()
	s.Available.Store(true)	
}
