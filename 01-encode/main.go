package main

import (
	"fmt"
)

const char = "0123456789abcdef"

func encode(src []byte) string {
	dst := make([]byte, len(src)*2)

	for i, p := range src {
		dst[i*2] = char[p>>4]
		dst[i*2+1] = char[p&0xf]
	}

	return string(dst)
}

// ascii => byte
func getByte(s byte) byte {
	for i, c := range char {
		if s == byte(c) {
			return byte(i)
		}
	}
	return 0
}

func decode(src string) []byte {
	dst := make([]byte, len(src)/2)

	for i := range dst {
		s0 := getByte(src[i*2])
		s1 := getByte(src[i*2+1])
		dst[i] = s0<<4 + s1
	}

	return dst
}

func main() {
	data := "hello, encoding :D"

	encoded := encode([]byte(data))

	fmt.Printf("original:\n%s\n\n", data)
	fmt.Printf("encode:\n%s\n\n", encoded)
	fmt.Printf("decode:\n%s\n", decode(encoded))
}
