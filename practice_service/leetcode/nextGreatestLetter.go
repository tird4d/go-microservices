package leetcode

func NextGreatestLetter(letters []byte, target byte) byte {
	//
	//target 	//d
	// l				   m 						h
	//'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'
	//'c', 'd', 'd', 'e', 'e', 'g', 'h', 'j', 'k', 'l'

	n := len(letters)
	h := n - 1
	l := 0

	// wrap
	if target >= letters[n-1] {
		return letters[0]
	}

	m := (l + h) / 2
	stop := false

	for stop != true {
		if letters[m] > target {
			h = m
			m = (l + h) / 2
			if h <= l {
				stop = true
				return letters[m]
			}

		} else {
			l = m + 1
			m = (h + l) / 2
		}
	}
	return letters[0]
}
