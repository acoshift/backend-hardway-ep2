package main

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h)
	mux.HandleFunc("/signout", signOut)
	http.ListenAndServe(":8080", mux)
}

func parseAuth(r *http.Request) (user string, pass string, ok bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", false
	}
	auth = strings.TrimPrefix(auth, "Basic ")

	ok = true

	userpass, _ := base64.StdEncoding.DecodeString(auth)
	if len(userpass) == 0 {
		return
	}

	sp := bytes.IndexByte(userpass, ':')
	user = string(userpass[:sp])
	pass = string(userpass[sp+1:])
	return
}

func h(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := parseAuth(r)
	if !ok {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if user != "superman" || subtle.ConstantTimeCompare([]byte(pass), []byte("hero")) != 1 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!doctype html>
		<h1>supersecret</h1>
		<a href=/signout>Sign Out</a>
	`))
}

func signOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`
		<!doctype html>
		<meta http-equiv=refresh content="0; url=/">
		<a href=/>Go to Home</a>
	`))
}
