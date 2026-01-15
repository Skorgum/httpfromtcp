package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/Skorgum/httpfromtcp/internal/response"
)

type Server struct {
	ln     net.Listener
	closed atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		ln: ln,
	}

	go srv.listen()

	return srv, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.ln.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	if err := response.WriteStatusLine(conn, response.StatusOk); err != nil {
		return
	}

	h := response.GetDefaultHeaders(0)

	if err := response.WriteHeaders(conn, h); err != nil {
		return
	}
}
