package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(c net.Conn) <-chan string {
	ch := make(chan string)
	go func() {
		defer c.Close()
		defer close(ch)
		line := ""
		for {
			buffer := make([]byte, 8)
			n, err := c.Read(buffer)
			if err != nil && !errors.Is(err, io.EOF) {
				log.Fatal(err)
			}
			data := string(buffer[:n])
			splitData := strings.Split(data, "\n")
			line += splitData[0] 
			if len(splitData) > 1 {
				ch <- line	
				line = splitData[1]
			}
			if errors.Is(err, io.EOF) {
				break
			}
		}	
	}()
	return ch
}

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

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
				
		fmt.Println("Channel Closed")
	}
}
