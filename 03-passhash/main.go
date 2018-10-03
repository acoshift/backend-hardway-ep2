package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// don't use in production
func hashPassword(password string) string {
	// generate salt
	salt := make([]byte, 16)
	rand.Read(salt)

	// hash password with salt
	h := sha256.New()
	h.Write(salt)
	hashed := h.Sum([]byte(password))

	hashedPassword := append([]byte{}, salt...)
	hashedPassword = append(hashedPassword, hashed...)
	return base64.RawStdEncoding.EncodeToString(hashedPassword)
}

func comparePassword(hashedPassword string, password string) bool {
	rawHashedPassword, _ := base64.RawStdEncoding.DecodeString(hashedPassword)

	salt := rawHashedPassword[:16]
	hashed := rawHashedPassword[16:]

	h := sha256.New()
	h.Write(salt)
	hashedNew := h.Sum([]byte(password))

	return subtle.ConstantTimeCompare(hashed, hashedNew) == 1
}

func main() {
	pass := "superman"
	fmt.Println("password:", pass)

	fmt.Println("---- custom algor ----")

	hashedCustom := hashPassword(pass)
	fmt.Println("hashed password:", hashedCustom)
	fmt.Println("hashed password again:", hashPassword(pass))

	fmt.Println("compare with 'supersaiyan':", comparePassword(hashedCustom, "supersaiyan"))
	fmt.Println("compare with 'superman':", comparePassword(hashedCustom, "superman"))

	// bcrypt
	fmt.Println("---- bcrypt ----")

	// encrypt OrpheanBeholderScryDoubt using blowfish with pass + salt as key
	hashedBcrypt, _ := bcrypt.GenerateFromPassword([]byte(pass), 10)
	fmt.Printf("hashed password: %s\n", hashedBcrypt)

	hashedBcrypt2, _ := bcrypt.GenerateFromPassword([]byte(pass), 10)
	fmt.Printf("hashed password again: %s\n", hashedBcrypt2)

	fmt.Println("compare with 'supersaiyan':", bcrypt.CompareHashAndPassword(hashedBcrypt, []byte("supersaiyan")) == nil)
	fmt.Println("compare with 'superman':", bcrypt.CompareHashAndPassword(hashedBcrypt, []byte("superman")) == nil)
}
