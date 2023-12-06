package main

import "fmt"

func main() {
	//fileRead, err := os.ReadFile("tests/big/posts.json")

	//fileRead, err := os.ReadFile("tests/test/pass1.json")
	//if err != nil {
	//	log.Fatal(err.Error())
	//}

	jsonStr := []byte(`{"e": 0.123456789e-12,
        "E": 1.234567890E+34,
        "":  23456789012E66,
        "zero": 0,
        "one": 1}`)

	l := JSONLexer{line: 1, column: 0}
	l.readJsonText(jsonStr)

	for {
		token, err := l.getNextToken()
		if err != nil {
			return
		}

		if token.name == "EOF" {
			break
		}

		fmt.Println(token)
	}
	//parser := NewJSONParser(jsonStr)
	//
	//obj, err := parser.parse()
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//
	//Printify(obj)

}
