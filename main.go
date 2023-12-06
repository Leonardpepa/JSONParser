package main

import (
	"encoding/json"
	"fmt"
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
	parser.useNumber()

	var obj interface{}
	obj, err = parser.parse()
	m := obj.(map[string]interface{})

	Printify(m)
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			fmt.Println(k, "is string", vv)
		case json.Number:
			fmt.Println(k, "is number", vv)
		case float64:
			fmt.Println(k, "is float64", vv)
		case []interface{}:
			fmt.Print(k, "is an array: ")
			fmt.Println(vv)
		case nil:
			fmt.Println(k, "is nil")
		case map[string]interface{}:
			fmt.Println(k, "is a map[string]interface{} ", vv)
		default:
			fmt.Println(k, "is of a type I don't know how to handle")
		}
	}
}
