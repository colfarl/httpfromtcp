// Package httpserver to serve as :w
package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/colfarl/httpfromtcp/internal/headers"
	"github.com/colfarl/httpfromtcp/internal/request"
	"github.com/colfarl/httpfromtcp/internal/response"
	"github.com/colfarl/httpfromtcp/internal/server"
)

const port = 42069
const badRequestHTML = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalErrHTML = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const okHTML = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func basicHandler(res *response.Writer, req *request.Request) {
	header := headers.NewHeaders()
	if req.RequestLine.RequestTarget == "/yourproblem" {
		res.WriteStatusLine(400)
		header.Set("Content-Type", "text/html")
		header.Set("Content-Length", strconv.Itoa(len(badRequestHTML)))
		header.Set("Connection", "close")
		res.WriteHeaders(header)
		res.WriteBody([]byte(badRequestHTML))
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		res.WriteStatusLine(500)
		header.Set("Content-Type", "text/html")
		header.Set("Content-Length", strconv.Itoa(len(internalErrHTML)))	
		header.Set("Connection", "close")
		res.WriteHeaders(header)
		res.WriteBody([]byte(internalErrHTML))
		return
	}
	
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		baseURL := "https://httpbin.org/" + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		resp, err := http.Get(baseURL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()	

		res.WriteStatusLine(200)
		header.Set("Content-Type", resp.Header.Get("Content-Type"))
		header.Set("Host", "httpbin.org")
		header.Set("Transfer-Encoding", "chunked")
		res.WriteHeaders(header)
		buf := make([]byte, 1024)
		total := make([]byte, 0)
		for {	
			n, err := resp.Body.Read(buf)
			if n > 0 {
				res.WriteChunkedBody(buf[:n])
				total = append(total, buf[:n]...)
			}

			if errors.Is(err, io.EOF) {
				res.WriteChunkedBodyDone()
				break
			}
			
			if err != nil {
				log.Fatal(err)
			}
		}

		hash := sha256.Sum256(total)
		hashHex := fmt.Sprintf("%x", hash)
		trailers := headers.NewHeaders()
		trailers.Set("X-Content-SHA256", string(hashHex))
		trailers.Set("X-Content-Length", strconv.Itoa(len(total)))
		res.WriteTrailers(trailers)
		return
	}

		
	if req.RequestLine.RequestTarget == "/video" {
		res.WriteStatusLine(200)
		header.Set("Content-Type", "video/mp4")
		video, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Fatal(err)
			return
		}

		header.Set("Content-Length", strconv.Itoa(len(video)))
		header.Set("Connection", "close")
		res.WriteHeaders(header)
		res.WriteBody([]byte(video))
		return
	}
	res.WriteStatusLine(200)
	header.Set("Content-Type", "text/html")
	header.Set("Content-Length", strconv.Itoa(len(okHTML)))
	res.WriteHeaders(header)
	res.WriteBody([]byte(okHTML))
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
