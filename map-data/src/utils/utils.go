package utils

import "fmt"

func IntToOrdinal(i int) string {
	if i < 0 {
		panic("Can't handle negative int")
	}

	suffixes := []string{"th", "st", "nd", "rd", "th", "th", "th", "th", "th", "th"}
	var suffix string
	if i%100 == 11 || i%100 == 12 || i%100 == 13 {
		suffix = suffixes[0]
	} else {
		suffix = suffixes[i%10]
	}
	return fmt.Sprintf("%v%v", i, suffix)
}
