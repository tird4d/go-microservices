package leetcode

import (
	"strings"
)

func RomanToInteger(s string) int {

	//  IV  = 4     A
	//  IX  = 9     B
	//  XL  = 40    E
	//  XC  = 90    F
	//  CD  = 400   G
	//  CM  = 900   K

	s = strings.ReplaceAll(s, "IV", "A")
	s = strings.ReplaceAll(s, "IX", "B")
	s = strings.ReplaceAll(s, "XL", "E")
	s = strings.ReplaceAll(s, "XC", "F")
	s = strings.ReplaceAll(s, "CD", "G")
	s = strings.ReplaceAll(s, "CM", "K")

	n := 0
	for _, char := range s {
		switch char {
		case 'I':
			n += 1
		case 'V':
			n += 5
		case 'X':
			n += 10
		case 'L':
			n += 50
		case 'C':
			n += 100
		case 'D':
			n += 500
		case 'M':
			n += 1000
		case 'A':
			n += 4
		case 'B':
			n += 9
		case 'E':
			n += 40
		case 'F':
			n += 90
		case 'G':
			n += 400
		case 'K':
			n += 900
		}
	}

	return n

}
