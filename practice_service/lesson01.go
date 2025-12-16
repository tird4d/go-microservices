package main

import "fmt"

func main() {
	// Hello World - First program
	fmt.Println("Welcome to Go Practice! 🎉")

	// ===================================
	// Lesson 1: Variables
	// ===================================

	// Method 1: Full declaration with var
	var name string = "Ali"

	// Method 2: Short declaration with :=
	age := 25

	// Method 3: Declaration without initial value (zero value)
	var score int // default value: 0

	fmt.Println("Name:", name)
	fmt.Println("Age:", age)
	fmt.Println("Score:", score)

	// ===================================
	// Exercise 1: Try it yourself!
	// ===================================
	// TODO: Define a variable of type float64 called price
	// TODO: Define a variable of type bool called isActive
	// TODO: Print their values

	var price float64
	var isActive bool
	price = 10.05
	isActive = true
	fmt.Println("Price:", price)

	fmt.Println("isActive", isActive)

}
