package debug

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPProfIndex(t *testing.T) {
	server := newDebugServer()
	defer server.Close()

	resp := get200(t, server.URL+"/debug/pprof/")

	if !strings.Contains(resp, "profiles") {
		t.Errorf("Unknown response:\n%s", resp)
	}
}

func TestPProfCmdline(t *testing.T) {
	server := newDebugServer()
	defer server.Close()

	resp := get200(t, server.URL+"/debug/pprof/cmdline")

	if !strings.Contains(resp, "http-handlers") {
		t.Errorf("Unknown response:\n%s", resp)
	}
}

func TestPProfSymbol(t *testing.T) {
	server := newDebugServer()
	defer server.Close()

	resp := get200(t, server.URL+"/debug/pprof/symbol")

	if !strings.Contains(resp, "num_symbols") {
		t.Errorf("Unknown response:\n%s", resp)
	}
}

func TestExpVars(t *testing.T) {
	server := newDebugServer()
	defer server.Close()

	resp := get200(t, server.URL+"/debug/vars")

	if !strings.Contains(resp, "memstats") {
		t.Errorf("Unknown response:\n%s", resp)
	}
}

func TestGCPostOnly(t *testing.T) {
	server := newDebugServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/debug/gc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 405 {
		t.Errorf("Status code was %d, but expected 405", resp.StatusCode)
	}
}

func TestGC(t *testing.T) {
	server := newDebugServer()
	defer server.Close()

	resp, err := http.Post(server.URL+"/debug/gc", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		t.Errorf("Status code was %d, but expected 200", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	actual := string(b)
	expected := "running GC...done!\n"
	if actual != expected {
		t.Errorf("Was %q, but expected %q", actual, expected)
	}
}

func get200(t *testing.T, url string) string {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		t.Errorf("Status code was %d, but expected 200", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func newDebugServer() *httptest.Server {
	return httptest.NewServer(Wrap(http.HandlerFunc(helloWorld)))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
