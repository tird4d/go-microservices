package main

import "fmt"

func main() {
	// Lesson 2: Control Flow

	// If-Else Example
	age := 20
	if age >= 18 {
		fmt.Println("You are an adult.")
	} else {
		fmt.Println("You are a minor.")
	}

	// Switch Example
	day := "Monday"
	switch day {
	case "Monday":
		fmt.Println("Start of the work week!")
	case "Friday":
		fmt.Println("Almost the weekend!")
	default:
		fmt.Println("Just another day.")
	}

	// For Loop Example
	for i := 1; i <= 5; i++ {
		fmt.Println("Iteration:", i)
	}

	// Exercise 2: Try it yourself!
	// TODO: Write a program that checks if a number is even or odd using if-else.
	// TODO: Use a switch statement to print a message for different grades (A, B, C, etc.).
	// TODO: Create a for loop that prints numbers from 10 to 1 in reverse order.
}
