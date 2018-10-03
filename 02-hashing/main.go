package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
)

func sum(b []byte) []byte {
	s := new(big.Int)
	for _, p := range b {
		s.Add(s, new(big.Int).SetInt64(int64(p)))
	}
	return s.Bytes()
}

func djb2(b []byte) uint64 {
	var hash uint64 = 5381
	for _, c := range b {
		hash = hash<<5 + hash + uint64(c)
	}
	return hash
}

func main() {
	data1 := "I do love you, don't I"
	data2 := "I don't love you, do I"
	fmt.Println("data1:", data1)
	fmt.Println("data2:", data2)

	sum1 := sum([]byte(data1))
	fmt.Printf("sum data1: %x\n", sum1)

	sum2 := sum([]byte(data2))
	fmt.Printf("sum data2: %x\n", sum2) // hash collision

	djb21 := djb2([]byte(data1))
	fmt.Printf("djb2 data1: %x\n", djb21)

	djb22 := djb2([]byte(data2))
	fmt.Printf("djb2 data2: %x\n", djb22)

	sha11 := sha1.Sum([]byte(data1))
	fmt.Printf("sha1 data1: %x\n", sha11)

	sha12 := sha1.Sum([]byte(data2))
	fmt.Printf("sha1 data2: %x\n", sha12)
}
