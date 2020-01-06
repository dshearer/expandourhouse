package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type state struct {
	Name string
	FIPS int
	USPS string
}

func printStringMap(out io.StringWriter, m map[string]int, name string) {
	out.WriteString(fmt.Sprintf("var %v = map[string]*State{\n", name))
	for k, v := range m {
		out.WriteString(fmt.Sprintf("    \"%v\": &All[%v],\n", k, v))
	}
	out.WriteString("}\n\n")
}

func printIntMap(out io.StringWriter, m map[int]int, name string) {
	out.WriteString(fmt.Sprintf("var %v = map[int]*State{\n", name))
	for k, v := range m {
		out.WriteString(fmt.Sprintf("    %v: &All[%v],\n", k, v))
	}
	out.WriteString("}\n\n")
}

func main() {
	// read states json
	f, err := os.Open("states.json")
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	var states []state
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&states); err != nil {
		log.Panic(err)
	}

	// organize
	byName := make(map[string]int)
	byUsps := make(map[string]int)
	byFips := make(map[int]int)
	for i, st := range states {
		byName[st.Name] = i
		byUsps[st.USPS] = i
		byFips[st.FIPS] = i
	}

	// open output file
	out, err := os.Create("states.go")
	if err != nil {
		log.Panic(err)
	}
	defer out.Close()

	// print package
	out.WriteString("package states\n\n")

	// print state struct
	out.WriteString(`type State struct {
	Name string
	Fips int
	Usps string
}`)
	out.WriteString("\n\n")

	// print array
	out.WriteString("var All = []State{\n")
	for _, st := range states {
		line := fmt.Sprintf("    State{Name: \"%v\", Fips: %v, Usps: \"%v\"},\n",
			st.Name, st.FIPS, st.USPS)
		out.WriteString(line)
	}
	out.WriteString("}\n\n")

	// print maps
	printStringMap(out, byName, "ByName")
	printStringMap(out, byUsps, "ByUsps")
	printIntMap(out, byFips, "ByFips")
}
