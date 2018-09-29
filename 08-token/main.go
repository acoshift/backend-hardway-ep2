package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h)
	mux.HandleFunc("/auth", auth)
	mux.HandleFunc("/signout", signOut)
	http.ListenAndServe(":8080", mux)
}

// token => user id
var tokenDatabase = map[string]int{}

func auth(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	if user != "superman" || pass != "hero" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := generateToken()
	tokenDatabase[token] = 1
	w.Write([]byte(token))
}

func generateToken() string {
	var b [32]byte
	rand.Read(b[:])
	return base64.RawStdEncoding.EncodeToString(b[:])
}

func parseToken(r *http.Request) string {
	// Authorization: Bearer TOKEN
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

func getUserID(token string) int {
	return tokenDatabase[token]
}

func h(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(parseToken(r))
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("supersecret cat"))
}

func signOut(w http.ResponseWriter, r *http.Request) {
	token := parseToken(r)
	delete(tokenDatabase, token)

	w.Write([]byte("okidoki"))
}
