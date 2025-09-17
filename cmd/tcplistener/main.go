package main

import (
	"fmt"
	"log"
	"net"

	"github.com/colfarl/httpfromtcp/internal/request"
)


func main() {

	l, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection Accepted")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", request.RequestLine.Method)
		fmt.Println("- Target:", request.RequestLine.RequestTarget)
		fmt.Println("- Version:", request.RequestLine.HttpVersion)
	

		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %v: %v\n", key, value)
		}
		
		fmt.Println("Body:")
		fmt.Println(string(request.Body))
		fmt.Println("Connection Closed")
	}
}

