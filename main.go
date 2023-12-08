package main

import (
	"encoding/json"
	"fmt"
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

func errorsFromInvalidJsonFiles() {
	for i := range make([]int, 33) {
		filename := fmt.Sprintf("tests/test/fail%d.json", i+1)
		input, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}

		var d interface{}
		err = json.Unmarshal(input, &d)
		if err != nil {
			fmt.Println(filename, ": ", err)
		}
	}
}
