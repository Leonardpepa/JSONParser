package main

import (
	"JSONParser/JSONParser"
	"JSONParser/Util"
	"log"
	"os"
)

func main() {
	input, err := os.ReadFile("tests/step4/valid.json")

	if err != nil {
		log.Fatal(err)
	}

	parsed, err := JSONParser.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	Util.Printify(parsed)
}
