package main

import (
	"log"
	"os"
)

func main() {
	fileRead, err := os.ReadFile("tests/big/posts.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	parser := NewParser(fileRead)
	err = parser.parse()

	if err != nil {
		log.Fatal(err.Error())
	}
}
