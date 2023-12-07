package main

import (
	"log"
	"os"
)

func main() {
	input, err := os.ReadFile("tests/test/pass2.json")
	if err != nil {
		log.Fatal(err)
	}
	parser := NewJSONParser(input)

	parsed, jsonError := parser.parse()

	if jsonError != nil {
		log.Fatal(jsonError)
	}

	Printify(parsed)

}
