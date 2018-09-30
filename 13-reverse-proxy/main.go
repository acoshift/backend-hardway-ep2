package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/acoshift/middleware"
)

func main() {
	srv := reverseProxy{
		Listen: 9000,
		Locations: []location{
			{"/", "https://www.google.com", true, false},
			{"/superman", "https://example.com", false, false},
			{"/nekopost/", "https://www.nekopost.net/", true, true},
		},
		GzipEnable: true,
	}
	log.Fatal(srv.Start())
}

type location struct {
	Path            string
	Target          string
	ReplaceBodyPath bool
	ForceCache      bool
}

type reverseProxy struct {
	*http.ServeMux
	transport    *http.Transport
	cacheStorage sync.Map // url => data

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

	gzipCompressor := middleware.GzipCompressor
	gzipCompressor.Skipper = func(*http.Request) bool {
		return !p.GzipEnable
	}

	h := middleware.Chain(
		middleware.Compress(gzipCompressor),
	)(p)

	return http.ListenAndServe(fmt.Sprintf(":%d", p.Listen), h)
}

type cacheItem struct {
	Header     http.Header
	Body       []byte
	StatusCode int
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

		// cache key = scheme://$host$requestUri
		cacheKey := r.URL.Scheme + "://" + r.Host + r.RequestURI

		// lookup, already have cache ?
		if data, ok := p.cacheStorage.Load(cacheKey); ok {
			item := data.(*cacheItem)

			copyHeader(w.Header(), item.Header)

			w.WriteHeader(item.StatusCode)

			// copy cache to w
			io.Copy(w, bytes.NewReader(item.Body))
			return
		}

		r.Header.Set("Referer", l.Target)

		resp, err := p.transport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		copyHeader(w.Header(), resp.Header)

		w.WriteHeader(resp.StatusCode)

		if l.ReplaceBodyPath && shouldReplaceBody(resp.Header.Get("Content-Type")) {
			var buf bytes.Buffer
			io.Copy(&buf, resp.Body)

			result := strings.Replace(buf.String(), l.Target, l.Path, -1)
			io.Copy(w, strings.NewReader(result))
			return
		}

		p.writeBody(w, resp.Body, cacheKey, l.ForceCache, resp.StatusCode)
	})
}

func copyHeader(dst, src http.Header) {
	for k, v := range src {
		for _, vv := range v {
			dst.Add(k, vv)
		}
	}
}

func (p *reverseProxy) writeBody(w http.ResponseWriter, r io.Reader, url string, forceCache bool, statusCode int) {
	if forceCache {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		h := make(http.Header)
		copyHeader(h, w.Header())
		h.Del("Content-Encoding")

		p.cacheStorage.Store(url, &cacheItem{
			StatusCode: statusCode,
			Body:       buf.Bytes(),
			Header:     h,
		})
		io.Copy(w, &buf)
		return
	}

	io.Copy(w, r)
}

func shouldReplaceBody(ct string) bool {
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
