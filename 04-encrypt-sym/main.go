package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

func main() {
	// generate key 32 bytes
	key := make([]byte, 32)
	rand.Read(key)
	fmt.Printf("key: %x\n", key)

	msg := []byte("hello, superman please save my cat")
	fmt.Printf("msg: %s\n", msg)

	// create new AES cipher block
	block, _ := aes.NewCipher(key)

	// generate nonce
	nonce := make([]byte, 12)
	rand.Read(nonce)
	fmt.Printf("nonce: %x\n", nonce)

	// create new GCM block operation
	aesgcm, _ := cipher.NewGCM(block)

	// encrypt
	ciphertext := aesgcm.Seal(nil, nonce, msg, nil)
	fmt.Printf("cipher text: %x\n", ciphertext)

	// decrypt
	plaintext, _ := aesgcm.Open(nil, nonce, ciphertext, nil)
	fmt.Printf("plain text: %s\n", plaintext)
}
