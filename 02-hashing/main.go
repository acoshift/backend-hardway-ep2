package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"math/big"
)

func main() {
	data, _ := ioutil.ReadFile("text.txt")

	sum := new(big.Int)
	for _, p := range data {
		sum.Add(sum, new(big.Int).SetInt64(int64(p)))
	}
	fmt.Printf("sum: %x\n", sum.Bytes())

	p := sha1.Sum(data)
	fmt.Printf("sha1: %x\n", p)
}
