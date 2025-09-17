// Package httpserver to serve as :w
package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/colfarl/httpfromtcp/internal/request"
	"github.com/colfarl/httpfromtcp/internal/response"
	"github.com/colfarl/httpfromtcp/internal/server"
)

const port = 42069

func basicHandler(w io.Writer, req *request.Request) *server.HandlerError{
	if req.RequestLine.RequestTarget == "/yourproblem"{
		return &server.HandlerError{
			StatusCode: response.BadRequest,
			Message: "Your problem is not my problem\n",
		}
	}

	if req.RequestLine.RequestTarget == "/myproblem"{
		return &server.HandlerError{
			StatusCode: response.InternalError,
			Message: "Woopsie, my bad\n",
		}
	}

	w.Write([]byte("All good, frfr\n"))
	return nil
}

func main() {
	server, err := server.Serve(port, basicHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
