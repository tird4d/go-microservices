package main

import (
	"fmt"
	"practice_service/leetcode"
)

func main() {
	// fmt.Println("Hello World")

	// fmt.Println(lessons.Lesson01())
	// lessons.Lesson02()
	// leetcode.IsPowerOfTwo(-16)
	// fmt.Println(leetcode.NextGreatestLetter([]byte{'c', 'd', 'd', 'e', 'e', 'g', 'h', 'j', 'k', 'l'}, 'j'))
	// fmt.Println(leetcode.RomanToInteger("MCMXCIV"))

	a := []int{1, 2, 3}
	b := leetcode.LinkListGenerator(a)

	fmt.Println(b)

	// result := leetcode.AddTwoNumbers(&a, &a)

	// fmt.Println(result)

}
