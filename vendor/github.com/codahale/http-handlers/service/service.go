// Package service combines the other various packages in http-handlers to
// provide an operations-friendly http.Handler for your application. Including
// this package will also allow you to dump a full stack trace to stderr by
// sending your application the SIGUSR1 signal.
package service

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/codahale/http-handlers/debug"
	"github.com/codahale/http-handlers/logging"
	"github.com/codahale/http-handlers/metrics"
	"github.com/codahale/http-handlers/recovery"
)

// Service is an HTTP service.
type Service struct {
	h *logging.LoggingHandler
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(w, r)
}

// Close stops the service.
func (s Service) Close() error {
	s.h.Stop()
	return nil
}

// New returns a new service-ready handler given an application handler.
//
// This stack application-level metrics, debug endpoints, panic recovery, and
// request logging, in that order.
func New(h http.Handler, onPanic recovery.PanicHandler) Service {
	l := logging.Wrap(
		recovery.Wrap(
			debug.Wrap(
				metrics.Wrap(
					h,
				),
			),
			onPanic,
		),
		os.Stdout,
	)
	l.Start()
	return Service{h: l}
}

func init() {
	dump := make(chan os.Signal)
	go func() {
		stack := make([]byte, 16*1024)
		for _ = range dump {
			n := runtime.Stack(stack, true)
			fmt.Fprintf(os.Stderr, "==== %s\n%s\n====\n", time.Now(), stack[0:n])
		}
	}()
	signal.Notify(dump, syscall.SIGUSR1)
}
