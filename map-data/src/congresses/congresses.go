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

func Get(n int) *Congress {
	if gCache == nil {
		gCache = make(map[int]*Congress)
	}

	congress, ok := gCache[n]
	if ok {
		return congress
	}

	if n <= 0 {
		panic("Invalid congress number")
	}

	congress = &Congress{
		Number:    n,
		Name:      fmt.Sprintf("%v Congress", utils.IntToOrdinal(n)),
		StartYear: gFirstCongressStartYear + 2*(n-1),
	}
	gCache[n] = congress
	return congress
}
