package main

import (
	"encoding/json"
	"fmt"
)

func Printify(j interface{}) {
	printWithIndent(j, 0)
	fmt.Print("\n\n")
}

func printWithIndent(j interface{}, indentationLevel int) {
	switch v := j.(type) {
	case map[string]interface{}:
		fmt.Println("{")
		i := 0
		for k, o := range v {
			printIndentation(indentationLevel + 1)
			fmt.Print("\""+k+"\"", ": ")
			printWithIndent(o, indentationLevel+1)
			if i == len(v)-1 {
				fmt.Println()
			} else {
				fmt.Println(",")
			}
			i++
		}
		printIndentation(indentationLevel + 1)
		fmt.Print("}")
	case []interface{}:
		fmt.Print("[")
		for index, o := range v {
			printWithIndent(o, indentationLevel+1)
			if index < len(v)-1 {
				fmt.Print(",")
			}
		}
		fmt.Print("]")
	case bool, float64, json.Number, nil:
		if v == nil {
			fmt.Println("null")
		} else {
			fmt.Print(v)
		}
	case string:
		fmt.Printf("%#v", v)
	default:
		fmt.Println("Unrecognisable type")
	}
}
func printIndentation(indentationLevel int) {
	for i := 0; i < indentationLevel; i++ {
		fmt.Print("  ") // You can adjust the number of spaces as needed
	}
}
