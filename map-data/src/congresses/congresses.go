// Package congresses contains helpers for getting info about sessions of the
// US Congress.
//
// A session begins after a Congressional election and ends just before the
// next one --- thus, each session lasts for two years.  Sessions are named
// like "First Congress", "Second Congress", "115th Congress", etc.  For
// n > 0, the nth Congress started in the year 1789 + 2*(n - 1).
package congresses

import (
	"fmt"

	"expandourhouse.com/mapdata/utils"
)

type Congress struct {
	Number    int
	Name      string
	StartYear int
}

var gCache map[int]*Congress

const gFirstCongressStartYear = 1789
const gLastCongress = 115

func GetAll() []*Congress {
	var array []*Congress
	for i := 1; i <= gLastCongress; i++ {
		array = append(array, Get(i))
	}
	return array
}

// Get returns info about the nth Congress.
func Get(n int) *Congress {
	if n <= 0 {
		panic("n must be positive")
	}

	if gCache == nil {
		gCache = make(map[int]*Congress)
	}

	congress, ok := gCache[n]
	if ok {
		return congress
	}

	congress = &Congress{
		Number:    n,
		Name:      fmt.Sprintf("%v Congress", utils.IntToOrdinal(n)),
		StartYear: gFirstCongressStartYear + 2*(n-1),
	}
	gCache[n] = congress
	return congress
}

// GetForYear returns info about the Congress that started in the given year
// (if there is one) or the Congress that was in session during the given year.
func GetForYear(year int) *Congress {
	if year < gFirstCongressStartYear {
		return nil
	}
	if year%2 == 0 {
		year--
	}
	n := (year-1789)/2 + 1
	return Get(n)
}
