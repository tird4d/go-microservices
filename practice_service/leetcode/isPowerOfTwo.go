package leetcode

import "fmt"

func IsPowerOfTwo(n int64) bool {
	fmt.Printf("this is the number %d \n", n)
	a := divideByTwo(n)
	if a == 0 {
		return true
	} else {
		return false
	}
}

func divideByTwo(n int64) int64 {
	a := n % 2
	if n == 2 || n == 1 {
		return 0
	}

	if a == 0 {
		b := n / 2
		c := divideByTwo(b)
		return c

	} else {
		return a
	}

}
