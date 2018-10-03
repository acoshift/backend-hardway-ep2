package main

import (
	"io"
	"log"
	"net"
	"net/http"

	"golang.org/x/sync/errgroup"
)

// HTTP Proxy
func main() {
	http.ListenAndServe(":9000", http.HandlerFunc(proxy))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleTunnel(w, r)
		return
	}

	handleHTTP(w, r)
}

func copy(dst io.Writer, src io.Reader) func() error {
	return func() error {
		_, err := io.Copy(dst, src)
		return err
	}
}

func handleTunnel(w http.ResponseWriter, r *http.Request) {
	log.Println("CONNECT", r.RequestURI)
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

	var eg errgroup.Group
	eg.Go(copy(upstream, client))
	eg.Go(copy(client, upstream))

	eg.Wait()
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)

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
