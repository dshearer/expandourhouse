package main

import (
	"encoding/base64"
	"io"
	"log"
	"os"
)

const gImports = `
package main

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

`

func main() {

	// open data file
	in, err := os.Open("tufts-all-votes-congress-3.tsv")
	if err != nil {
		log.Panic(err)
	}
	defer in.Close()

	// open output file
	out, err := os.Create("data.go")
	if err != nil {
		log.Panic(err)
	}
	defer out.Close()

	// write imports
	if _, err := out.WriteString(gImports); err != nil {
		log.Panic(err)
	}

	// write variable decl
	if _, err := out.WriteString("const gTuftsDataBase64 = `"); err != nil {
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

	// finish decode func
	if _, err := out.WriteString(gDecodeFunc); err != nil {
		log.Panic(err)
	}
}
