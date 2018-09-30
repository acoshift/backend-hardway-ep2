package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	srv := reverseProxy{
		Listen: 9000,
		Locations: []location{
			{"/", "https://www.google.com", true},
			{"/superman", "https://example.com", false},
			{"/nekopost/", "https://www.nekopost.net/", true},
		},
	}
	log.Fatal(srv.Start())
}

type location struct {
	Path            string
	Target          string
	ReplaceBodyPath bool
}

type reverseProxy struct {
	*http.ServeMux
	transport *http.Transport

	Listen     int
	Locations  []location
	GzipEnable bool
}

func (p *reverseProxy) Start() error {
	// setup
	p.transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	p.ServeMux = http.NewServeMux()

	for _, l := range p.Locations {
		p.ServeMux.Handle(l.Path, p.makeReverseProxy(&l))
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", p.Listen), p)
}

func (p *reverseProxy) makeReverseProxy(l *location) http.Handler {
	targetURL, _ := url.Parse(l.Target)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = targetURL.Scheme
		r.URL.Host = targetURL.Host
		r.URL.Path = strings.TrimPrefix(r.URL.Path, l.Path)
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
		r.Host = targetURL.Host
		r.Header.Set("Referer", l.Target)

		resp, err := p.transport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		// already compress ?
		if resp.Header.Get("Content-Encoding") == "" && shouldCompress(resp.Header.Get("Content-Type")) {
			// not compress

			// check is browser support gzip ?
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				gw := &gzipRW{
					ResponseWriter: w,
				}
				gw.Init()
				defer gw.Close()

				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Del("Content-Length")

				w = gw
			}
		}

		w.WriteHeader(resp.StatusCode)

		if l.ReplaceBodyPath && shouldCompress(resp.Header.Get("Content-Type")) {
			var buf bytes.Buffer
			io.Copy(&buf, resp.Body)

			result := strings.Replace(buf.String(), l.Target, l.Path, -1)
			io.Copy(w, strings.NewReader(result))
			return
		}

		io.Copy(w, resp.Body)
	})
}

type gzipRW struct {
	http.ResponseWriter

	gzipWriter *gzip.Writer
}

func (w *gzipRW) Init() {
	w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
}

func (w *gzipRW) Close() {
	w.gzipWriter.Close()
}

func (w *gzipRW) Write(p []byte) (int, error) {
	return w.gzipWriter.Write(p)
}

func shouldCompress(ct string) bool {
	mt, _, _ := mime.ParseMediaType(ct)

	switch mt {
	case "text/html":
	case "text/plain":
	case "application/json":
	case "application/javascript":
	case "application/x-javascript":
	case "text/javascript":
	case "text/css":
	default:
		return false
	}
	return true
}
