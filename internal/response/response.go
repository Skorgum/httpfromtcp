package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Skorgum/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string

	switch statusCode {
	case StatusOk:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	default:
		reason = ""
	}

	if reason == "" {
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n", int(statusCode))
		return err
	}
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", int(statusCode), reason)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for key, value := range h {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", key, value); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "\r\n"); err != nil {
		return err
	}

	return nil
}
