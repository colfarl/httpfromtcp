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

type Writer struct {
	Buffer		io.Writer
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.NewHeaders()
	defaultHeaders["content-length"] = strconv.Itoa(contentLen)
	defaultHeaders["connection"] = "close"
	defaultHeaders["content-type"] = "text/plain"
	return defaultHeaders
}

func NewWriter(w io.Writer) Writer{
	return Writer{
		Buffer: w,
	}
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	w.Buffer.Write(p)
	return len(p), nil
}

func (w * Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case OK:
		w.Buffer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return nil
	case BadRequest:
		w.Buffer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return nil
	case InternalError:
		w.Buffer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return nil
	default:
		return fmt.Errorf("unknown status code")
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {

	headerBytes := make([]byte, 0)
	for key, value := range headers {
		headerBytes = fmt.Appendf(headerBytes, "%s: %s\r\n", key, value)
	}

	headerBytes = fmt.Append(headerBytes, "\r\n")
	w.Buffer.Write(headerBytes)

	return nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error){
	head := []byte(fmt.Sprintf("%x\r\n", len(p)))

    // payload + \r\n
    b := make([]byte, 0, len(head)+len(p)+2)
    b = append(b, head...)
    b = append(b, p...)
    b = append(b, '\r', '\n')

    return w.WriteBody(b)
}

func (w *Writer) WriteChunkedBodyDone() (int, error){
	n, err := w.WriteBody([]byte("0\r\n"))
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	return w.WriteHeaders(h)
}
