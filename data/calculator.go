package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println("Hello from the calculator!")

	result := add(10, 5)
	fmt.Printf("10 + 5 = %d\n", result)

	result = subtract(10, 5)
	fmt.Printf("10 - 5 = %d\n", result)

	result = multiply(10, 5)
	fmt.Printf("10 * 5 = %d\n", result)

	quotient := divide(10, 5)
	fmt.Printf("10 / 5 = %.2f\n", quotient)

	fmt.Printf("Square root of 16 = %.2f\n", math.Sqrt(16))
}

func add(a, b int) int {
	return a + b
}

func subtract(a, b int) int {
	return a - b
}

func multiply(a, b int) int {
	return a * b
}

func divide(a, b float64) float64 {
	return a / b
}
