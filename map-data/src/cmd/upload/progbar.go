package main

import "fmt"

type ProgBar struct {
	Total   int
	current int
}

func (self *ProgBar) AddProgress(n int) {
	self.SetProgress(self.current + n)
}

func (self *ProgBar) SetProgress(n int) {
	if n == self.current {
		return
	}

	if n > self.Total {
		self.current = self.Total
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
	nbrStars := int((float32(self.current) / float32(self.Total)) * float32(totalPlaces))
	nbrSpaces := totalPlaces - nbrStars
	percent := 100 * (float32(self.current) / float32(self.Total))

	fmt.Print("\r[")
	for i := 0; i < nbrStars; i++ {
		fmt.Print("*")
	}
	for i := 0; i < nbrSpaces; i++ {
		fmt.Print(" ")
	}
	fmt.Printf(" ] %3.2f%% (%d/%d)", percent, self.current, self.Total)
}
