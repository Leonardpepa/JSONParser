package main

import (
	"log"
	"os"
)

func main() {
	//fileRead, err := os.ReadFile("tests/big/posts.json")

	fileRead, err := os.ReadFile("tests/step4/valid2.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	parser := NewParser(fileRead)

	obj, err := parser.parse()
	if err != nil {
		log.Fatal(err.Error())
	}

	Printify(obj)

}
