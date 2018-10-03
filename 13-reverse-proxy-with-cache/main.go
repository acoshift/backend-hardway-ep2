package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

func main() {
	http.ListenAndServe(":9000", http.HandlerFunc(nekopostReverseProxy))
}

type cacheItem struct {
	Body   []byte
	Header http.Header
}

var cacheStorage sync.Map

func scheme(r *http.Request) string {
	if r.TLS == nil {
		return "http"
	}
	return "https"
}

func nekopostReverseProxy(w http.ResponseWriter, r *http.Request) {
	cacheKey := scheme(r) + "://" + r.Host + r.RequestURI

	// has cached data ?
	if it, ok := cacheStorage.Load(cacheKey); ok {
		fmt.Println(r.RequestURI, "- HIT")
		cache := it.(*cacheItem)

		copyHeader(w.Header(), cache.Header)
		w.WriteHeader(http.StatusOK)
		io.Copy(w, bytes.NewReader(cache.Body))
		return
	}

	fmt.Println(r.RequestURI, "- MISS")

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

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	if shouldReplaceData(resp.Header.Get("Content-Type")) {
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)

		re := regexp.MustCompile("http(s)?://(www\\.)?nekopost.net")

		result := re.ReplaceAllString(buf.String(), "http://localhost:9000")

		if resp.StatusCode == http.StatusOK {
			storeCache(cacheKey, w.Header(), []byte(result))
		}

		io.Copy(w, strings.NewReader(result))
		return
	}

	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)

	// cache only 200
	if resp.StatusCode == http.StatusOK {
		storeCache(cacheKey, w.Header(), buf.Bytes())
	}

	io.Copy(w, &buf)
}

func storeCache(key string, header http.Header, body []byte) {
	cloneHeader := make(http.Header)
	copyHeader(cloneHeader, header)
	cacheStorage.Store(key, &cacheItem{
		Header: cloneHeader,
		Body:   body,
	})
}

func copyHeader(dst http.Header, src http.Header) {
	for k, v := range src {
		for _, vv := range v {
			dst.Add(k, vv)
		}
	}
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
