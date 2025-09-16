// Package request for http request
package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	State		int	
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ErrMalformedRequestLine = errors.New("request line has incorrect format")
var ErrInvalidMethod = errors.New("invalid method in request line")
var ErrUnsupportedHTTPVersion = errors.New("unsupported http version in request line; supported versions [HTTP/1.1]")

var ValidMethod = map[string]struct{}{
	"GET": {}, 
	"POST" : {}, 
	"PUT" : {},
	"DELETE" : {}, 
	"PATCH" : {}, 
}
const crlf = "\r\n"
const bufferSize = 8

func requestLineFromString(rawRequestLine string) (*RequestLine, error) {
	res := strings.Split(rawRequestLine, " ")
	if res == nil || len(res) != 3 {
		return nil, ErrMalformedRequestLine 
	}

	method := res[0]
	if _, ok := ValidMethod[method]; !ok {
		return nil, ErrInvalidMethod
	}

	version := res[2]
	if version != "HTTP/1.1" {
		return nil, ErrUnsupportedHTTPVersion
	}
	
	return &RequestLine{
		HttpVersion: "1.1",
		RequestTarget: res[1],
		Method: method,
	},	nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error){
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, len(data), nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State  {
	case 0:
		requestLine, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		} else if bytesRead == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.State = 1
		return bytesRead, nil
	case 1:
		return 0, errors.New("trying to read in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}


func resizeBuffer(buffer []byte) ([]byte) {
	if len(buffer) == cap(buffer){
		newBuffer := make([]byte, cap(buffer) * 2)
		copy(newBuffer, buffer)
		return newBuffer
	}
	return buffer
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0

	request := Request{
		State: 0,
	}
	
	for request.State != 1 {
		buffer = resizeBuffer(buffer)	
		bytesRead, err := reader.Read(buffer[readToIndex:])
		if errors.Is(err, io.EOF) {
			request.State = 1
			break
		} else if err != nil {
			return nil, err	
		}

		readToIndex += bytesRead
		bytesWritten, err := request.parse(buffer)
		if err != nil {
			return nil, err
		}
		if bytesWritten > 0 {
			newBuffer := make([]byte, len(buffer))
			copy(newBuffer, buffer[bytesWritten:])
			readToIndex -= bytesWritten
			buffer = newBuffer
		}
	}
	return &request, nil
}
