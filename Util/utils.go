package Util

import (
	"encoding/json"
	"fmt"
)

func Printify(object interface{}) {
	printWithIndent(object, 0)
	fmt.Print("\n\n")
}

func printMap(object map[string]interface{}, indentationLevel int) {
	fmt.Println("{")
	i := 0
	for k, o := range object {
		printIndentation(indentationLevel + 1)
		fmt.Print("\""+k+"\"", ": ")
		printWithIndent(o, indentationLevel+1)
		if i == len(object)-1 {
			fmt.Println()
		} else {
			fmt.Println(",")
		}
		i++
	}
	printIndentation(indentationLevel)
	fmt.Print("}")
}

func printArray(array []interface{}, indentationLevel int) {
	fmt.Print("[")
	for index, o := range array {
		printWithIndent(o, indentationLevel+1)
		if index < len(array)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Print("]")
}

func printWithIndent(object interface{}, indentationLevel int) {
	switch v := object.(type) {
	case map[string]interface{}:
		printMap(v, indentationLevel)
	case []interface{}:
		printArray(v, indentationLevel)
	case bool, float64, json.Number, nil:
		if v == nil {
			fmt.Print("null")
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
