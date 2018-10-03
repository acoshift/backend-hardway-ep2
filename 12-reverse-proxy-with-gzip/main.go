package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func main() {
	http.ListenAndServe(":9000", http.HandlerFunc(nekopostReverseProxy))
}

func nekopostReverseProxy(w http.ResponseWriter, r *http.Request) {
	r.URL.Scheme = "https"
	r.URL.Host = "www.nekopost.net"
	r.Host = "www.nekopost.net" // host is http header "Host"

	refURL, _ := url.Parse(r.Referer())
	if refURL != nil {
		r.Header.Set("Referer", "https://www.nekopost.net"+refURL.Path)
	}

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

	ct := resp.Header.Get("Content-Type")

	// skip compressed response
	if resp.Header.Get("Content-Encoding") == "" && shouldCompress(ct) {
		// browser support gzip ?
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

	if shouldReplaceData(ct) {
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)

		re := regexp.MustCompile("http(s)?://(www\\.)?nekopost.net")

		result := re.ReplaceAllString(buf.String(), "http://localhost:9000")

		io.Copy(w, strings.NewReader(result))
		return
	}

	io.Copy(w, resp.Body)
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

func shouldReplaceData(ct string) bool {
	mt, _, _ := mime.ParseMediaType(ct)

	switch mt {
	case "text/html":
	case "application/javascript":
	case "application/x-javascript":
	case "text/javascript":
	case "text/css":
	default:
		return false
	}
	return true
}
