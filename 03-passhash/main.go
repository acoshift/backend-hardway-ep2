package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pass := "superman"

	// encrypt OrpheanBeholderScryDoubt with pass + salt
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), 10)

	err := bcrypt.CompareHashAndPassword(hash, []byte("superman"))
	if err != nil {
		fmt.Println("pass not match")
		return
	}
	fmt.Println("pass matched")
}
