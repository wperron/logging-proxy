package main

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testHandler struct{}

func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

type gzipHandler struct{}

func (gh *gzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Encoding", "gzip")
	w.Header().Add("Content-Lenght", "32")
	w.Write([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0xfe, 0x4f, 0x0e, 0x62,
		0x00, 0x03, 0xf3, 0x48, 0xcd, 0xc9, 0xc9, 0xd7,
		0x51, 0x08, 0xcf, 0x2f, 0xca, 0x49, 0x01, 0x00,
		0xc6, 0x86, 0x5b, 0x26, 0x0c, 0x00, 0x00, 0x00,
	}) // "Hello, World" gzipped
}

func TestOK(t *testing.T) {
	downstream := httptest.NewServer(new(testHandler))

	req, err := http.NewRequest("GET", downstream.URL, nil)
	if err != nil {
		panic(err)
	}

	lt := &LoggingTripper{next: http.DefaultTransport}

	_, err = lt.RoundTrip(req)
	if err != nil {
		t.Errorf("executing HTTP request: %s", err)
		t.FailNow()
	}
}

type errTripper struct{}

func (et *errTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("test error")
}

func TestErr(t *testing.T) {
	lt := &LoggingTripper{next: &errTripper{}}
	req, err := http.NewRequest("GET", "http://example.org", nil)
	if err != nil {
		panic(err)
	}

	resp, err := lt.RoundTrip(req)
	if err == nil {
		t.Error("expected request to fail, got nil error")
	} else {
		if resp != nil {
			t.Error("expected response to be nil in case of an error")
		}
	}
}

func TestHeaders(t *testing.T) {
	downstream := httptest.NewServer(new(gzipHandler))

	req, err := http.NewRequest("GET", downstream.URL, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Accept-Encoding", "gzip")
	lt := &LoggingTripper{next: http.DefaultTransport}

	resp, err := lt.RoundTrip(req)
	if err != nil {
		t.Errorf("executing HTTP request: %s", err)
		t.FailNow()
	}

	encod := resp.Header.Get("Content-Encoding")
	if encod != "gzip" {
		t.Errorf("expected gzip encoding, got: \"%s\"", encod)
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(err)
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	if string(raw) != "Hello, World" {
		t.Errorf("expected hello world message, got: %s", string(raw))
	}
}

func TestNewRoundTripper(t *testing.T) {
	u, _ := url.Parse("http://example.org")
	lt := NewLoggingReverseProxy(u)
	if lt == nil {
		t.Errorf("failed to create instance of RoundTripper: got nil")
	}

	if _, ok := lt.(http.Handler); !ok {
		t.Errorf("expected implementation of http.Handler, got: %T", lt)
	}
}
