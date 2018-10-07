package main

import (
	"io"
	"log"
	"net"
	"net/http"
)

// HTTP Proxy
func main() {
	http.ListenAndServe(":9000", http.HandlerFunc(proxy))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)

	if r.Method == http.MethodConnect {
		handleTunnel(w, r)
		return
	}

	handleHTTP(w, r)
}

func handleTunnel(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Proxy not support hijacker", http.StatusInternalServerError)
		return
	}

	// dial to upstream
	upstream, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer upstream.Close()

	w.WriteHeader(http.StatusOK)

	client, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	go io.Copy(upstream, client)
	io.Copy(client, upstream)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
