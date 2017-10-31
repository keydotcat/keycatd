// Package recovery provides an HTTP handler which recovers panics in an
// underlying handler, logs debug information about the panic, and returns a 500
// Internal Server Error to the client.
package recovery

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"

	"github.com/codahale/metrics"
)

// PanicHandler is a handler for per-request panics.
type PanicHandler func(id int64, err interface{}, stacktrace []string, request *http.Request)

// LogOnPanic logs the given panic and its stacktrace, prefixing each line with
// the panic ID.
func LogOnPanic(id int64, err interface{}, stacktrace []string, _ *http.Request) {
	logMutex.Lock()
	defer logMutex.Unlock()

	log.Printf("panic=%016x message=%v\n", id, err)
	for _, line := range stacktrace {
		log.Printf("panic=%016x %s", id, line)
	}
}

var logMutex sync.Mutex

// Wrap returns an handler which proxies requests to the given handler, but
// handles panics by logging the stack trace and returning a 500 Internal Server
// Error to the client, if possible.
func Wrap(h http.Handler, onPanic PanicHandler) http.Handler {
	return &recoveryHandler{h: h, p: onPanic}
}

type recoveryHandler struct {
	h http.Handler
	r sync.Mutex
	p PanicHandler
}

func (h *recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		e := recover()
		if e != nil {
			panics.Add()

			id := rand.Int63()
			var lines []string
			for skip := 1; ; skip++ {
				pc, file, line, ok := runtime.Caller(skip)
				if !ok {
					break
				}
				if file[len(file)-1] == 'c' {
					continue
				}
				f := runtime.FuncForPC(pc)
				s := fmt.Sprintf("%s:%d %s()\n", file, line, f.Name())
				lines = append(lines, s)
			}
			h.p(id, e, lines, r)

			body := fmt.Sprintf(
				"%s\n%016x",
				http.StatusText(http.StatusInternalServerError),
				id,
			)
			http.Error(w, body, http.StatusInternalServerError)

		}
	}()
	h.h.ServeHTTP(w, r)
}

var panics = metrics.Counter("HTTP.Panics")
