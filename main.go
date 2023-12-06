package main

import (
	"log"
	"os"
)

func main() {
	//fileRead, err := os.ReadFile("tests/big/posts.json")

	fileRead, err := os.ReadFile("tests/step4/valid.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	parser := NewParser(fileRead)

	var obj interface{}
	obj, err = parser.parse()
	m := obj.(map[string]interface{})
	Printify(m)
}
