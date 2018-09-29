package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

// ECDSA-SHA256
func main() {
	fmt.Println("algor: ECDSA-SHA256")

	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	fmt.Println("private key:", base64.StdEncoding.EncodeToString(priv.D.Bytes()))

	msg := []byte("hello, superman please save my dog.")
	fmt.Println("msg:", string(msg))

	hash := sha256.Sum256(msg)
	fmt.Println("hash msg (sha256):", hex.EncodeToString(hash[:]))

	r, s, _ := ecdsa.Sign(rand.Reader, priv, hash[:])
	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)
	fmt.Println("signature:", base64.StdEncoding.EncodeToString(signature))

	pubKey := priv.X.Bytes()
	pubKey = append(pubKey, priv.Y.Bytes()...)
	fmt.Println("public key:", base64.StdEncoding.EncodeToString(pubKey))

	// load public key
	pub := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(pubKey[:32]),
		Y:     new(big.Int).SetBytes(pubKey[32:]),
	}

	// load signature
	r = new(big.Int).SetBytes(signature[:32])
	s = new(big.Int).SetBytes(signature[32:])

	ok := ecdsa.Verify(pub, hash[:], r, s)
	fmt.Println("verify result:", ok)
}
