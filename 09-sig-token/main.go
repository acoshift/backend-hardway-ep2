package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var hmacKey = []byte("keyboard cat")

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h)
	mux.HandleFunc("/auth", auth)
	http.ListenAndServe(":8080", mux)
}

func auth(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	if user != "superman" || pass != "hero" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(&token{
		ID: 1,
	}, hmacKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(token))
}

func generateToken(t *token, key []byte) (string, error) {
	// jwt: meta.data.sig

	// Gob Token
	// data.sig
	// data => gob => base64

	rand.Read(t.Nonce[:])
	t.ExpiresAt = time.Now().Add(time.Minute).Unix()

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(t)
	if err != nil {
		return "", err
	}

	sig := signHMAC(buf.Bytes(), key)

	// data + "." + sig => token
	token := base64.RawURLEncoding.EncodeToString(buf.Bytes())
	token += "."
	token += base64.RawURLEncoding.EncodeToString(sig)
	return token, nil
}

func signHMAC(data []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	return h.Sum(data)
}

type token struct {
	ID        int
	ExpiresAt int64
	Nonce     [6]byte
}

func parseToken(r *http.Request, key []byte) *token {
	// Authorization: Bearer TOKEN
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil
	}

	tk := strings.TrimPrefix(auth, "Bearer ")
	// data.sig
	sp := strings.Index(tk, ".")
	data, _ := base64.RawURLEncoding.DecodeString(tk[:sp])
	sig, _ := base64.RawURLEncoding.DecodeString(tk[sp+1:])

	newSig := signHMAC(data, key)
	if !bytes.Equal(sig, newSig) {
		return nil
	}

	var tokenData token
	gob.NewDecoder(bytes.NewReader(data)).Decode(&tokenData)
	if time.Now().Unix() > tokenData.ExpiresAt {
		return nil
	}

	return &tokenData
}

func h(w http.ResponseWriter, r *http.Request) {
	tokenData := parseToken(r, hmacKey)
	if tokenData == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "supersecret cat for user %d", tokenData.ID)
}
