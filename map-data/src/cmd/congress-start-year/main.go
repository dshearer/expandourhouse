package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"expandourhouse.com/mapdata/congresses"
)

const gUSAGE = "usage: congressStartYear CONGRESS_NBR\n"

func main() {
	// get args
	flag.Parse()
	if flag.NArg() != 1 {
		os.Stderr.WriteString(gUSAGE)
		os.Exit(1)
	}
	nbrStr := flag.Arg(0)
	nbr, err := strconv.Atoi(nbrStr)
	if err != nil {
		os.Stderr.WriteString(gUSAGE)
		os.Exit(1)
	}
	if nbr <= 0 {
		os.Stderr.WriteString("Congress number must be > 0\n")
		os.Exit(1)
	}

	// print start year
	fmt.Printf("%v\n", congresses.Get(nbr).StartYear)
}
