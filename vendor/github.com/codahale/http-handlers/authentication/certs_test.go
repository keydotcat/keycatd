package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testHeaderName = "Client-Subject-DN"
	testOU         = "MyOU"
	testDN         = "/C=foo/OU=MyOU"
)

func TestX509NameVerifier(t *testing.T) {
	v := X509NameVerifier{
		HeaderName:       testHeaderName,
		CheckCertificate: RequireOU([]string{testOU}),
	}

	handler := v.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	// Missing header
	res := httptest.ResponseRecorder{}
	handler.ServeHTTP(&res, &http.Request{})
	if res.Code != 403 {
		t.Errorf("Expected to receive a 403 when missing a cert header, got %d", res.Code)
	}

	// Correct header
	hdr := http.Header{}

	hdr.Set(testHeaderName, testDN)

	res = httptest.ResponseRecorder{}
	handler.ServeHTTP(&res, &http.Request{Header: hdr})
	if res.Code != 204 {
		t.Errorf("Expected to receive a 204 with the correct OU, got %d", res.Code)
	}

	// Incorrect header
	hdr.Set(testHeaderName, "/C=foo/OU=another")

	res = httptest.ResponseRecorder{}
	handler.ServeHTTP(&res, &http.Request{Header: hdr})
	if res.Code != 403 {
		t.Errorf("Expected to receive a 403 with the wrong OU, got %d", res.Code)
	}

	// Invalid header
	hdr.Set(testHeaderName, "foo")

	res = httptest.ResponseRecorder{}
	handler.ServeHTTP(&res, &http.Request{Header: hdr})
	if res.Code != 403 {
		t.Errorf("Expected to receive a 403 with the wrong OU, got %d", res.Code)
	}

	// With a custom handler
	v.InvalidHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	res = httptest.ResponseRecorder{}
	handler.ServeHTTP(&res, &http.Request{Header: hdr})
	if res.Code != 404 {
		t.Errorf("Expected to receive a 404 from our custom handler, got %d", res.Code)
	}
}
