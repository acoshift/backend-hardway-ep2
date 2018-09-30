package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

/*
openssl genrsa -out server.key 2048
openssl rsa -in server.key -out server.key
openssl req -new -key server.key -days 3650 -out server.crt -x509
*/

// openssl ecparam -name prime256v1 -genkey -noout -out key.pem
// openssl req -new -key key.pem -days 3650 -out server.crt -x509

// HTTP Proxy
func main() {
	http.ListenAndServe(":9000", http.HandlerFunc(proxy))

	// srv := http.Server{
	// 	Addr:         ":9443",
	// 	Handler:      http.HandlerFunc(proxy),
	// 	TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	// }

	// err := srv.ListenAndServeTLS("server.crt", "server.key")
	// log.Fatal(err)
}

var transport = &http.Transport{
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

func proxy(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	resp, err := transport.RoundTrip(r)
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
