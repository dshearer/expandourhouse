package main

import (
	"io"
	"log"
	"os"
)

func main() {

	in, err := os.Open("style-template.json")
	if err != nil {
		log.Panic(err)
	}
	defer in.Close()

	out, err := os.Create("styleTemplate.go")
	if err != nil {
		log.Panic(err)
	}
	defer out.Close()
	if _, err := out.WriteString("package main\n\nconst gStyleTemplate = `"); err != nil {
		log.Panic(err)
	}
	if _, err := io.Copy(out, in); err != nil {
		log.Panic(err)
	}
	if _, err := out.WriteString("`\n"); err != nil {
		log.Panic(err)
	}
}
