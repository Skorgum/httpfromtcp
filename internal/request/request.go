package request

import (
	"fmt"
	"io"
	"strings"

	"github.com/Skorgum/httpfromtcp/internal/headers"
)

type parseState int

const (
	stateInitialized parseState = iota
	stateParsingHeaders
	stateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       parseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{
		state:   stateInitialized,
		Headers: headers.NewHeaders(),
	}

	for req.state != stateDone {
		// grow buffer if full
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if n > 0 {
			readToIndex += n

			consumed, parseErr := req.parse(buf[:readToIndex])
			if parseErr != nil {
				return nil, parseErr
			}

			if consumed > 0 {
				copy(buf, buf[consumed:readToIndex])
				readToIndex -= consumed
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	if req.state != stateDone {
		return nil, fmt.Errorf("incomplete request")
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	requestString := string(data)

	idx := strings.Index(requestString, "\r\n")
	if idx < 0 {
		return nil, 0, nil
	}

	parts := strings.Split(requestString[:idx], " ")
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("Invalid format")
	}

	version := strings.Split(parts[2], "/")
	if len(version) != 2 {
		return nil, 0, fmt.Errorf("Invalid format")
	}

	if version[1] != "1.1" {
		return nil, 0, fmt.Errorf("Unsupported HTML type")
	}

	for _, c := range parts[0] {
		if c < 'A' || c > 'Z' {
			return nil, 0, fmt.Errorf("Invalid format")
		}
	}

	newRequestLine := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version[1],
	}

	bytesConsumed := idx + len("\r\n")
	return newRequestLine, bytesConsumed, nil
}

func (r *Request) parse(data []byte) (int, error) {
	total := 0

	for r.state != stateDone {
		n, err := r.parseSingle(data[total:])
		if err != nil {
			return 0, err
		}
		total += n
		if n == 0 {
			break
		}
	}

	return total, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case stateInitialized:
		r1, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *r1
		r.state = stateParsingHeaders
		return n, nil

	case stateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = stateDone
		}
		return n, nil

	case stateDone:
		return 0, fmt.Errorf("cannot parse in done state")

	default:
		return 0, fmt.Errorf("unknown state")
	}
}
