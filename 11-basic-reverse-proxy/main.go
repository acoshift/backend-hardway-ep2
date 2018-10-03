package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	srv := reverseProxy{
		Listen: 9000,
		Locations: []location{
			{"/", "https://www.google.com"},
			{"/superman", "https://example.com"},
		},
	}
	log.Fatal(srv.Start())
}

type location struct {
	Path   string
	Target string
}

type reverseProxy struct {
	*http.ServeMux

	Listen    int
	Locations []location
}

func (p *reverseProxy) Start() error {
	// setup
	p.ServeMux = http.NewServeMux()

	for _, l := range p.Locations {
		p.ServeMux.Handle(l.Path, p.makeReverseProxy(l.Target))
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", p.Listen), p)
}

func (reverseProxy) makeReverseProxy(target string) http.Handler {
	targetURL, _ := url.Parse(target)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = targetURL.Scheme
		r.URL.Host = targetURL.Host
		r.Host = targetURL.Host

		resp, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// copy header
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(resp.StatusCode)

		io.Copy(w, resp.Body)
	})
}
