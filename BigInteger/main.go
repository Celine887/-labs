package main

import (
	"fmt"
)

func main() {
	fmt.Println("ITMO-Endian Uint239 Example")

	a := uint239.FromUint32(42, 0)
	b := uint239.FromUint32(10, 0)

	sum := uint239.Add(a, b)
	fmt.Printf("42 + 10 = %d\n", extractValue(sum))

	diff := uint239.Subtract(a, b)
	fmt.Printf("42 - 10 = %d\n", extractValue(diff))

	prod := uint239.Multiply(a, b)
	fmt.Printf("42 * 10 = %d\n", extractValue(prod))

	quot := uint239.Divide(a, b)
	fmt.Printf("42 / 10 = %d\n", extractValue(quot))
}

func extractValue(n uint239.Uint239) uint32 {
	var val uint32

	for i := 0; i < len(n.Data); i++ {
		val = (val << 7) | uint32(n.Data[i]&0x7F)
	}

	return val
}
