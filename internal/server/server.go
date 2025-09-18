// Package server used to accept requests and send responses
package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/colfarl/httpfromtcp/internal/request"
	"github.com/colfarl/httpfromtcp/internal/response"
)
type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode	response.StatusCode
	Message		string
}
type Server struct {
	Available	*atomic.Bool
	Listener	net.Listener
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
		log.Print(err)
		return
	}
	res := response.NewWriter(conn) 
	handler(&res, r)	
	conn.Close()
}
