package response

import (
	"fmt"
	"io"

	"github.com/Skorgum/httpfromtcp/internal/headers"
)

type writerState int

const (
	stateInit writerState = iota
	stateStatusWritten
	stateHeadersWritten
	stateBodyWritten
	stateTrailersWritten
)

type Writer struct {
	state writerState
	w     io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state: stateInit,
		w:     w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateInit {
		return fmt.Errorf("cannot write status line in state %d", w.state)
	}

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

	var err error
	if reason == "" {
		_, err = fmt.Fprintf(w.w, "HTTP/1.1 %d \r\n", int(statusCode))
	} else {
		_, err = fmt.Fprintf(w.w, "HTTP/1.1 %d %s\r\n", int(statusCode), reason)
	}
	if err != nil {
		return err
	}

	w.state = stateStatusWritten
	return nil
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != stateStatusWritten {
		return fmt.Errorf("cannot write headers in state %d", w.state)
	}

	for k, v := range h {
		line := []byte(fmt.Sprintf("%s: %s\r\n", k, v))
		if _, err := w.w.Write(line); err != nil {
			return err
		}
	}

	if _, err := w.w.Write([]byte("\r\n")); err != nil {
		return err
	}

	w.state = stateHeadersWritten
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, fmt.Errorf("cannot write body in state %d", w.state)
	}

	w.state = stateBodyWritten
	return w.w.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten && w.state != stateBodyWritten {
		return 0, fmt.Errorf("cannot write chunked body in state %d", w.state)
	}

	chunkSize := len(p)
	total := 0

	n, err := fmt.Fprintf(w.w, "%x\r\n", chunkSize)
	if err != nil {
		return total, err
	}
	total += n

	n, err = w.w.Write(p)
	if err != nil {
		return total, err
	}
	total += n

	n, err = w.w.Write([]byte("\r\n"))
	if err != nil {
		return total, err
	}
	total += n

	w.state = stateBodyWritten
	return total, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateBodyWritten && w.state != stateHeadersWritten {
		return 0, fmt.Errorf("cannot finish chunked body in state %d", w.state)
	}

	n, err := w.w.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != stateBodyWritten {
		return fmt.Errorf("cannot write trailers in state %d", w.state)
	}

	for k, v := range h {
		line := []byte(fmt.Sprintf("%s: %s\r\n", k, v))
		if _, err := w.w.Write(line); err != nil {
			return err
		}
	}

	if _, err := w.w.Write([]byte("\r\n")); err != nil {
		return err
	}

	w.state = stateTrailersWritten

	return nil
}
