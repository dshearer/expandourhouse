package main

import "fmt"

type ProgBar struct {
	total   int
	current int
}

func NewProgBar(total int) *ProgBar {
	return &ProgBar{total: total}
}

func (self *ProgBar) SetProgress(n int) {
	if n == self.current {
		return
	}

	if n > self.total {
		self.current = self.total
	} else {
		self.current = n
	}
	self.print()
}

func (self *ProgBar) Reset() {
	self.current = 0
}

func (self *ProgBar) print() {
	totalPlaces := 100
	nbrStars := int((float32(self.current) / float32(self.total)) * float32(totalPlaces))
	nbrSpaces := totalPlaces - nbrStars
	percent := 100 * (float32(self.current) / float32(self.total))

	fmt.Print("\r[")
	for i := 0; i < nbrStars; i++ {
		fmt.Print("*")
	}
	for i := 0; i < nbrSpaces; i++ {
		fmt.Print(" ")
	}
	fmt.Printf(" ] %3.2f%% (%d/%d)", percent, self.current, self.total)
}
