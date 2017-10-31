package metrics

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codahale/metrics"
)

func TestMetrics(t *testing.T) {
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(w, "hello, world")
	}))
	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := http.Get(s.URL + "/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Status code was %d, but expected 200", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	a := string(b)
	e := "hello, world\n"
	if a != e {
		t.Errorf("Response was %q, but expected %q", a, e)
	}

	counters, gauges := metrics.Snapshot()

	expectedGauges := []string{
		"HTTP.Latency.P50",
		"HTTP.Latency.P75",
		"HTTP.Latency.P90",
		"HTTP.Latency.P95",
		"HTTP.Latency.P99",
		"HTTP.Latency.P999",
	}

	for _, name := range expectedGauges {
		if _, ok := gauges[name]; !ok {
			t.Errorf("Missing gauge %q", name)
		}
	}

	expectedCounters := []string{
		"HTTP.Requests",
		"HTTP.Responses",
	}

	for _, name := range expectedCounters {
		if _, ok := counters[name]; !ok {
			t.Errorf("Missing counter %q", name)
		}
	}
}

func BenchmarkMetrics(b *testing.B) {
	var (
		r *http.Request
		w http.ResponseWriter
		h = Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.ServeHTTP(w, r)
		}
	})
}
