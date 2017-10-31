// Package metrics provides an HTTP handler which registers counters for the
// number of requests received and responses sent as well as quantiles of the
// latency of responses.
package metrics

import (
	"net/http"
	"time"

	"github.com/codahale/metrics"
)

// Wrap returns a handler which records the number of requests received and
// responses sent to the given handler, as well as latency quantiles for
// responses over a five-minute window.
//
// These counters are published as the following metrics:
//
//     HTTP.Requests
//     HTTP.Responses
//     HTTP.Latency.{P50,P75,P90,P95,P99,P999}
//
// By tracking incoming requests and outgoing responses, one can monitor not
// only the requests per second, but also the number of requests being processed
// at any given point in time.
func Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add()                  // inc requests
		defer responses.Add()           // inc responses when we're done
		defer recordLatency(time.Now()) // record latency when we're done

		h.ServeHTTP(w, r)
	})
}

var (
	requests  = metrics.Counter("HTTP.Requests")
	responses = metrics.Counter("HTTP.Responses")

	// a five-minute window tracking 1ms-3min
	latency = metrics.NewHistogram("HTTP.Latency", 1, 1000*60*3, 3)
)

func recordLatency(start time.Time) {
	elapsedMS := time.Now().Sub(start).Seconds() * 1000.0
	_ = latency.RecordValue(int64(elapsedMS))
}
