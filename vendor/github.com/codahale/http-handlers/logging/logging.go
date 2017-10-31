// Package logging provides a fast, asynchronous request logger which outputs
// NCSA/Apache combined logs.
package logging

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// A LoggingHandler is a HTTP handler which proxies requests to an underlying
// handler and logs the results.
type LoggingHandler struct {
	clock   clock
	w       io.Writer
	handler http.Handler
	buffer  chan string
	quit    chan struct{}
}

// Wrap returns the underlying handler, wrapped in a LoggingHandler which will
// write to the given Writer. N.B.: You must call Start() on the result before
// using it.
func Wrap(h http.Handler, w io.Writer) *LoggingHandler {
	return &LoggingHandler{
		clock:   time.Now,
		w:       w,
		handler: h,
		buffer:  make(chan string, 1000),
		quit:    make(chan struct{}),
	}
}

// Start creates a goroutine to handle the logging IO.
func (al *LoggingHandler) Start() {
	go func() {
		for s := range al.buffer {
			fmt.Fprint(al.w, s)
		}
		close(al.quit)
	}()
}

// Stop closes the internal channel used to buffer log statements and waits for
// the IO goroutine to complete.
func (al *LoggingHandler) Stop() {
	close(al.buffer)
	<-al.quit
}

func (al *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wrapper := &responseWrapper{w: w, status: 200}

	start := al.clock()
	al.handler.ServeHTTP(wrapper, r)
	end := al.clock()

	remoteAddr := r.RemoteAddr
	if index := strings.LastIndex(remoteAddr, ":"); index != -1 {
		remoteAddr = remoteAddr[:index]
	}

	if s := r.Header.Get(xForwardedFor); s != "" {
		remoteAddr = s
	}

	referer := r.Referer()
	if "" == referer {
		referer = "-"
	}

	userAgent := r.UserAgent()
	if "" == userAgent {
		userAgent = "-"
	}

	al.buffer <- fmt.Sprintf(
		"%s %s %s [%s] \"%s %s %s\" %d %d %q %q %d %q\n",
		remoteAddr,
		"-", // We're not supporting identd, sorry.
		"-", // We're also not supporting basic auth.
		start.In(time.UTC).Format(apacheFormat),
		r.Method,
		r.RequestURI,
		r.Proto,
		wrapper.status,
		0,
		referer,
		userAgent,
		end.Sub(start).Nanoseconds()/int64(time.Millisecond),
		r.Header.Get(xRequestID),
	)
}

const (
	apacheFormat  = "02/Jan/2006:15:04:05 -0700"
	xRequestID    = "X-Request-Id"
	xForwardedFor = "X-Forwarded-For"
)

type responseWrapper struct {
	w      http.ResponseWriter
	status int
}

func (w *responseWrapper) Header() http.Header {
	return w.w.Header()
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *responseWrapper) WriteHeader(status int) {
	w.status = status
	w.w.WriteHeader(status)
}

func (w *responseWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.w.(http.Hijacker); ok {
		return hijacker.Hijack()
	} else {
		return nil, nil, errors.New("http-handler: wrapped responsewrapper does not implement http.Hijack")
	}
}

type clock func() time.Time
