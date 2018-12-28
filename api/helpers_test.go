package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
)

var gCli *http.Client
var activeSessionToken string
var activeCsrfToken string

func CheckErrorAndResponse(t *testing.T, r *http.Response, err error, code int) {
	if err != nil {
		t.Fatalf("Unpexpected error when %s %s: %s", r.Request.Method, r.Request.URL, err)
	}
	if r.StatusCode != code {
		contents := &bytes.Buffer{}
		contents.ReadFrom(r.Body)
		debug.PrintStack()
		t.Fatalf("Unexpected response code when %s %s: %d %s vs %d", r.Request.Method, r.Request.URL, r.StatusCode, contents, code)
	}
}

func CheckResponse(t *testing.T, r *http.Response, code int, keys ...string) string {
	defer r.Body.Close()
	if r.StatusCode != code {
		d := &bytes.Buffer{}
		d.ReadFrom(r.Body)
		t.Fatalf("Unexpected http code %d (expected 200) Data: %s", r.StatusCode, d.String())
	}
	if header, ok := r.Header["Content-Type"]; !ok {
		t.Error("No content-type header")
	} else {
		ct, _, err := mime.ParseMediaType(header[0])
		if err != nil {
			t.Errorf("Could not parse content-type: %s", err)
		} else if ct != "application/json" {
			t.Errorf("Unexpected content type: %s", ct)
		}
	}
	m := make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		t.Fatalf("Could not decode json: %s", err)
	}
	sm := m
	skeys := keys
	for len(skeys) > 1 {
		ism, ok := sm[skeys[0]]
		if !ok {
			t.Fatalf("Could not get %v from %v", keys, m)
		}
		sm, ok = ism.(map[string]interface{})
		if !ok {
			t.Fatalf("Could not get %v from %v", keys, m)
		}
		skeys = skeys[1:]
	}
	if len(skeys) == 1 {
		iv, ok := sm[skeys[0]]
		if !ok {
			t.Fatalf("Expected key %s is not there %v", keys, m)
		}
		v, ok := iv.(string)
		if !ok {
			t.Fatalf("Value is not string %s in %v", keys, m)
		}
		return v
	}
	return ""
}

func PostRequestWithHeader(path string, obj interface{}, header http.Header) (*http.Response, error) {
	body, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", srv.URL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.Header.Add("Content-Type", "application/json")
	return httpDo(req)
}

func PostRequestRaw(path string, data []byte, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest("POST", srv.URL+path, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))
	return httpDo(req)
}

func PostRequest(path string, obj interface{}) (*http.Response, error) {
	body, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", srv.URL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{}
	req.Header.Add("Content-Type", "application/json")
	return httpDo(req)
}

func PatchRequest(path string, obj interface{}) (*http.Response, error) {
	body, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PATCH", srv.URL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{}
	req.Header.Add("Content-Type", "application/json")
	return httpDo(req)
}

func GetRequest(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", srv.URL+path, nil)
	if err != nil {
		return nil, err
	}
	return httpDo(req)
}

func DeleteRequestWithBody(path string, obj interface{}) (*http.Response, error) {
	body, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("DELETE", srv.URL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	return httpDo(req)
}

func DeleteRequest(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", srv.URL+path, nil)
	if err != nil {
		return nil, err
	}
	return httpDo(req)
}

func PostForm(path string, v url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", srv.URL+path, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return httpDo(req)
}

func EventRequest(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", srv.URL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	return httpDo(req)
}

func getCookieJar() http.CookieJar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	if len(activeCsrfToken) == 0 {
		activeCsrfToken = "dummy"
	}
	val, err := apiH.csrf.sc.Encode(CSRF_COOKIE_NAME, activeCsrfToken)
	if err != nil {
		panic(err)
	}
	u, err := url.Parse(srv.URL)
	if err != nil {
		panic(err)
	}
	jar.SetCookies(u, []*http.Cookie{&http.Cookie{Name: CSRF_COOKIE_NAME, Value: val, Path: "/"}})
	return jar
}

func httpDo(req *http.Request) (*http.Response, error) {
	if gCli == nil {
		gCli = &http.Client{}
		gCli.Jar = getCookieJar()
	}
	if req.Header == nil {
		req.Header = http.Header{}
	}
	req.Header.Add("X-Csrf-Token", activeCsrfToken)
	if len(activeSessionToken) != 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", activeSessionToken))
	}
	return gCli.Do(req)
}
