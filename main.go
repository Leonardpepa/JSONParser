package main

import (
	"log"
	"os"
)

func main() {
	input, err := os.ReadFile("tests/step4/valid.json")

	if err != nil {
		log.Fatal(err)
	}

	p := JSONParser{}
	parsed, err := p.parse(input)

	if err != nil {
		log.Fatal(err)
	}

	Printify(parsed)
}
