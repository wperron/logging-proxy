package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	addr  = flag.String("addr", ":3202", "address to listen on")
	proxy = flag.String("proxy", "", "target server to proxy")
)

func main() {
	flag.Parse()

	target, err := url.Parse(*proxy)
	if err != nil {
		log.Fatal(err)
	}

	handler := NewLoggingReverseProxy(target)

	log.Println(fmt.Sprintf(
		"Proxy server started at %s, forwarding requests to %s", *addr, target,
	))
	log.Fatal(http.ListenAndServe(*addr, handler))
}

func NewLoggingReverseProxy(target *url.URL) http.Handler {
	handler := httputil.NewSingleHostReverseProxy(target)
	handler.Transport = &LoggingTripper{next: http.DefaultTransport}
	return handler
}

type LoggingTripper struct {
	next http.RoundTripper
}

func (lt *LoggingTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := lt.next.RoundTrip(r)
	if err != nil {
		log.Println(fmt.Sprintf("unexpected error: %s", err))
		return nil, err
	}

	log.Println(resp.StatusCode, r.Method, r.URL)
	r.Write(log.Default().Writer())
	return resp, nil
}
