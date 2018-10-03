package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
)

func startMockServer(addr string) {
	http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Serve from %s", addr)
	}))
}

func main() {
	servers := []string{
		"127.0.0.1:8001",
		"127.0.0.1:8002",
		"127.0.0.1:8003",
		"127.0.0.1:8004",
		"127.0.0.1:8005",
	}

	for _, addr := range servers {
		go startMockServer(addr)
	}

	var index int64

	p := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			// round-robin
			i := atomic.AddInt64(&index, 1)
			if i >= int64(len(servers)) {
				atomic.StoreInt64(&index, 0)
				i = 0
			}

			upstream := servers[i]

			r.URL.Host = upstream
			r.URL.Scheme = "http"
		},
	}
	http.ListenAndServe(":9000", p)
}
