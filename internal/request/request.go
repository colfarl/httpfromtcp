// Package request for http request
package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/colfarl/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers		headers.Headers
	Body		[]byte
	State		int //0 -> initialized; 1 -> parsed requestline, 2 -> parsed headers; 3 -> Done
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
	return requestLine, len(data[:idx]) + 2, nil
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
		bytesRead, finished, err := r.Headers.Parse(data) 	 
		if err != nil {
			return 0, err
		} else if bytesRead == 0 {
			return 0, nil
		}

		if finished {
			r.State = 2
		}

		return bytesRead, nil
	case 2:
		r.Body = append(r.Body, data...)
		return len(data), nil
	default:
		return 0, errors.New("unknown state")
	}
}


func resizeBuffer(buffer []byte, idx int) ([]byte) {
	if idx == len(buffer){
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
		Headers: headers.NewHeaders(),
		Body: make([]byte, 0),
	}
	
	// Request Line
	for request.State != 1 {
		buffer = resizeBuffer(buffer, readToIndex)	
		bytesRead, err := reader.Read(buffer[readToIndex:])

		if bytesRead > 0 {
			readToIndex += bytesRead
			bytesWritten, perr := request.parse(buffer[:readToIndex])
			if perr != nil {
				return nil, perr
			}

			if bytesWritten > 0 {
				copy(buffer, buffer[bytesWritten:readToIndex])
				readToIndex -= bytesWritten
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF){
				if readToIndex > 0 {
					bytesWritten, perr := request.parse(buffer[:readToIndex])
					if perr != nil {
						return nil, perr
					}
					if bytesWritten > 0 {
						copy(buffer, buffer[bytesWritten:readToIndex])
						readToIndex -= bytesWritten
					}
				}
				if request.State != 1 {
					return nil, io.ErrUnexpectedEOF
				}	
				break
			}
			return nil, err
		}	
	}
	
	// Request Headers
	for request.State != 2 {
		buffer = resizeBuffer(buffer, readToIndex)
		bytesRead, err := reader.Read(buffer[readToIndex:])

		if bytesRead > 0 {
			readToIndex += bytesRead
			bytesWritten, perr := request.parse(buffer[:readToIndex])
			if perr != nil {
				return nil, perr
			}

			if bytesWritten > 0 {
				copy(buffer, buffer[bytesWritten:readToIndex])
				readToIndex -= bytesWritten
			}
		}

		if err != nil {
			if readToIndex > 0 {
				bytesWritten, perr := request.parse(buffer[:readToIndex])
				if perr != nil {
					return nil, perr
				}
				if bytesWritten > 0 {
					copy(buffer, buffer[bytesWritten:readToIndex])
					readToIndex -= bytesWritten
				}
			}
			if errors.Is(err, io.EOF){
				if request.State != 2 {
					return nil, io.ErrUnexpectedEOF
				}
				break
			}
			return nil, err
		}	
	}
	

	length, ok := request.Headers.Get("Content-Length")
	if !ok {
		request.State = 3
	}
	
	intLength, err := strconv.Atoi(length)
	if err != nil && ok {
		return nil, fmt.Errorf("improper Content-Length format")
	}
	for request.State != 3 {
		buffer = resizeBuffer(buffer, readToIndex)
		bytesRead, err := reader.Read(buffer[readToIndex:])
		if bytesRead > 0 {
			readToIndex += bytesRead
			bytesWritten, perr := request.parse(buffer[:readToIndex])
			if perr != nil {
				return nil, perr
			}

			if bytesWritten > 0 {
				copy(buffer, buffer[bytesWritten:readToIndex])
				readToIndex -= bytesWritten
			}
		}

		if err != nil {
			if readToIndex > 0 {
				bytesWritten, perr := request.parse(buffer[:readToIndex])
				if perr != nil {
					return nil, perr
				}
				if bytesWritten > 0 {
					copy(buffer, buffer[bytesWritten:readToIndex])
					readToIndex -= bytesWritten
				}
			}
			if errors.Is(err, io.EOF) {
				if len(request.Body) != intLength {
					return nil, io.ErrUnexpectedEOF
				}
				break
			}
			return nil, err
		}
		
		
		if len(request.Body) == intLength {
			break
		}
		if len(request.Body) > intLength {
			return nil, fmt.Errorf("mismatched length and body")
		}
	}

	if len(request.Body) != intLength {
		return nil, fmt.Errorf("mismatched length and body")
	}
	request.State = 3
	return &request, nil
}
