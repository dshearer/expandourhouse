package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
)

const gImports = `
package turnout

import (
	"encoding/base64"
	"bytes"
	"io"
)

`

const gDecodeFunc = `
func OpenTuftsData() io.ReadSeeker {
	data, _ := base64.StdEncoding.DecodeString(gTuftsDataBase64)
	return bytes.NewReader(data)
}

func OpenHarvardData() io.ReadSeeker {
	data, _ := base64.StdEncoding.DecodeString(gHarvardDataBase64)
	return bytes.NewReader(data)
}

`

func writeDataVar(varName string, in io.Reader, out *os.File) {
	// write variable decl
	if _, err := out.WriteString(fmt.Sprintf("const %v = `", varName)); err != nil {
		log.Panic(err)
	}

	// write data as base64
	encoder := base64.NewEncoder(base64.StdEncoding, out)
	if _, err := io.Copy(encoder, in); err != nil {
		log.Panic(err)
	}
	if err := encoder.Close(); err != nil {
		log.Panic(err)
	}

	// finish var decl
	if _, err := out.WriteString("`\n"); err != nil {
		log.Panic(err)
	}
}

func main() {

	// open data files
	tuftsIn, err := os.Open("tufts-all-votes-congress-3.tsv")
	if err != nil {
		panic(err)
	}
	defer tuftsIn.Close()
	harvardIn, err := os.Open("harvard-1976-2018-house2.csv")
	if err != nil {
		panic(err)
	}
	defer harvardIn.Close()

	// open output file
	out, err := os.Create("turnoutData.go")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// write imports
	if _, err := out.WriteString(gImports); err != nil {
		panic(err)
	}

	// write variables
	writeDataVar("gTuftsDataBase64", tuftsIn, out)
	writeDataVar("gHarvardDataBase64", harvardIn, out)

	// finish decode func
	if _, err := out.WriteString(gDecodeFunc); err != nil {
		panic(err)
	}
}
