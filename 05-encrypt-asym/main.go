package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

func main() {
	// generate rsa private key
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	x509Priv := x509.MarshalPKCS1PrivateKey(priv)
	fmt.Printf("private key (x509):\n%s\n\n", base64.StdEncoding.EncodeToString(x509Priv))

	// encode to pem format
	keyPemBlock := pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509Priv}
	keyPem := pem.EncodeToMemory(&keyPemBlock)
	fmt.Printf("private key (pem):\n%s\n\n", keyPem)

	x509Pub, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubPemBlock := pem.Block{Type: "PUBLIC KEY", Bytes: x509Pub}
	pubPem := pem.EncodeToMemory(&pubPemBlock)
	fmt.Printf("public key (pem):\n%s\n\n", pubPem)

	msg := []byte("hello, superman save the world please.")
	fmt.Printf("msg: %s\n", msg)

	ciphertext, _ := rsa.EncryptPKCS1v15(rand.Reader, &priv.PublicKey, msg)
	fmt.Printf("cipher text: %x\n", ciphertext)

	plaintext, _ := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	fmt.Printf("plain text: %s\n", plaintext)
}
