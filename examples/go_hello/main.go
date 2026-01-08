package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("Hello from Go!")
	fmt.Printf("GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Number of CPUs: %d\n", runtime.NumCPU())

	// Simulate some work
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	fmt.Printf("Computed sum: %d\n", sum)
}
