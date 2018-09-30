package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

func main() {
	servers := []string{
		"https://www.google.com",
		"https://acourse.io",
		"https://github.com",
		"https://www.nekopost.net",
	}

	var index int64

	p := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			// round-robin

			i := atomic.LoadInt64(&index)
			i++
			if i >= int64(len(servers)) {
				i = 0
			}
			atomic.StoreInt64(&index, i)

			target := servers[i]
			targetURL, _ := url.Parse(target)

			r.URL.Host = targetURL.Host
			r.URL.Scheme = targetURL.Scheme
			r.Host = targetURL.Host
		},
	}
	http.ListenAndServe(":9000", p)
}
