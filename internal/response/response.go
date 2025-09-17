// Package response for formatting http responses
package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/colfarl/httpfromtcp/internal/headers"
)

type StatusCode int
const (
	OK				StatusCode = 200
	BadRequest		StatusCode = 400
	InternalError	StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case OK:
		w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return nil
	case BadRequest:
		w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return nil
	case InternalError:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return nil
	default:
		return fmt.Errorf("unknown status code")
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.NewHeaders()
	defaultHeaders["Content-Length"] = strconv.Itoa(contentLen)
	defaultHeaders["Connection"] = "close"
	defaultHeaders["Content-Type"] = "text/plain"
	return defaultHeaders
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {

	headerBytes := make([]byte, 0)
	for key, value := range headers {
		headerBytes = fmt.Appendf(headerBytes, "%s: %s\r\n", key, value)
	}

	headerBytes = fmt.Append(headerBytes, "\r\n")
	w.Write(headerBytes)

	return nil
}
