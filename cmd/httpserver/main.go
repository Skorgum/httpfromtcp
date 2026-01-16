package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Skorgum/httpfromtcp/internal/headers"
	"github.com/Skorgum/httpfromtcp/internal/request"
	"github.com/Skorgum/httpfromtcp/internal/response"
	"github.com/Skorgum/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget

	if strings.HasPrefix(target, "/httpbin/") {
		path := strings.TrimPrefix(target, "/httpbin")
		upstreamURL := "https://httpbin.org" + path

		res, err := http.Get(upstreamURL)
		if err != nil {
			log.Println("Something went wrong:", err)
			return
		}
		defer res.Body.Close()

		w.WriteStatusLine(response.StatusOk)

		h := response.GetDefaultHeaders(0)
		delete(h, "Content-Length")
		h.Override("Transfer-Encoding", "chunked")
		h.Override("Content-Type", res.Header.Get("Content-Type"))
		h.Override("Trailer", "X-Content-SHA256, X-Content-Length")

		w.WriteHeaders(h)

		buf := make([]byte, 1024)
		var fullBody []byte

		for {
			n, err := res.Body.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				fullBody = append(fullBody, chunk...)

				if _, werr := w.WriteChunkedBody(chunk); werr != nil {
					log.Println("error writing chunk:", werr)
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("error reading upstream:", err)
				return
			}
		}

		//log.Println("fullBody length:", len(fullBody)) //Testing

		if _, err := w.WriteChunkedBodyDone(); err != nil {
			log.Println("error finishing chunked body:", err)
			return
		}

		sum := sha256.Sum256(fullBody)
		hashHex := hex.EncodeToString(sum[:])
		lenghStr := strconv.Itoa(len(fullBody))

		trailers := headers.Headers{
			"X-Content-SHA256": hashHex,
			"X-Content-Length": lenghStr,
		}

		//log.Println("hashHex:", hashHex, "lenStr:", lenghStr) //Testing
		//log.Printf("trailers: %#v\n", trailers)               //Testing

		if err := w.WriteTrailers(trailers); err != nil {
			log.Panicln("error writing trailers:", err)
			return
		}
	}

	switch target {
	case "/yourproblem":
		body := []byte(`<html>
	<head>
		<title>400 Bad Request</title>
	</head>
	<body>
		<h1>Bad Request</h1>
		<p>Your request honestly kinda sucked.</p>
	</body>
</html>
`)

		w.WriteStatusLine(response.StatusBadRequest)

		h := response.GetDefaultHeaders(len(body))
		h.Override("Content-Type", "text/html")

		w.WriteHeaders(h)
		w.WriteBody(body)

	case "/myproblem":
		body := []byte(`
		<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`)

		w.WriteStatusLine(response.StatusInternalServerError)

		h := response.GetDefaultHeaders(len(body))
		h.Override("Content-Type", "text/html")

		w.WriteHeaders(h)
		w.WriteBody(body)

	default:
		body := []byte(`
		<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`)

		w.WriteStatusLine(response.StatusOk)

		h := response.GetDefaultHeaders(len(body))
		h.Override("Content-Type", "text/html")

		w.WriteHeaders(h)
		w.WriteBody(body)
	}
}
