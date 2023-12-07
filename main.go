package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	input, err := os.ReadFile("tests/test/pass1.json")
	if err != nil {
		log.Fatal(err)
	}
	lexer := JSONLexer{line: 1, column: 0}
	lexer.readJsonText(input)

	d := json.NewDecoder(bytes.NewReader(input))
	printTokens(lexer, d)
}

func printTokens(lexer JSONLexer, d *json.Decoder) {
	for {
		token, err := lexer.getNextToken()

		if err != nil {
			log.Fatal(err.Error())
		}

		if token.name == "Colon" || token.name == "Comma" {
			continue
		}

		goken, goerror := d.Token()

		if token.name == "EOF" && goerror == io.EOF {
			break
		} else if token.name == "EOF" || goerror == io.EOF {
			log.Fatal("ERROR lexers didint end together")
		}

		if goerror != nil {
			log.Println(goerror.Error())
		}
		if fmt.Sprintf("%v", token.value) == fmt.Sprintf("%v", goken) {
			log.Println(token.value, goken)
		}
	}
}
