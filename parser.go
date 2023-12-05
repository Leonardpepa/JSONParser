package main

import (
	"fmt"
	"log"
	"strconv"
)

type Parser struct {
	lexer         *JSONLexer
	lookahead     Token
	errorOccurred error
}

func NewParser(jsonBytes []byte) Parser {
	parser := Parser{}
	parser.lexer = &JSONLexer{column: 0, line: 1}
	parser.lexer.readJsonText(jsonBytes)
	return parser
}

func (parser *Parser) match(tType string) Token {
	if parser.lookahead.name == tType {
		nextToken, err := parser.lexer.getNextToken()
		if err != nil {
			return Token{}
		}
		prev := parser.lookahead
		parser.lookahead = nextToken
		return prev
	}
	return Token{}
}

func (parser *Parser) parse() error {
	nextT, err := parser.lexer.getNextToken()
	if err != nil {
		return err
	}

	parser.lookahead = nextT

	parsedJson := parser.parseValue()

	if parser.lookahead.value == "EOF" {
		funcName(parsedJson, 0)
		return nil
	} else {
		return fmt.Errorf("an Error occurred")
	}
}

func funcName(j JSONValue, indentationLevel int) {
	switch v := j.(type) {
	case JSONObject:
		fmt.Println("{")
		i := 0
		for k, o := range v.members {
			printIndentation(indentationLevel + 1)
			fmt.Print("\""+k+"\"", ": ")
			funcName(o, indentationLevel+1)
			if i == len(v.members)-1 {
				fmt.Println()
			} else {
				fmt.Println(",")
			}
			i++
		}
		printIndentation(indentationLevel + 1)
		fmt.Print("}")
	case JSONArray:
		fmt.Print("[")
		for index, o := range v.elements {
			funcName(o, indentationLevel+1)
			if index < len(v.elements)-1 {
				fmt.Print(",")
			}
		}
		fmt.Print("]")
	case JSONNull:
		fmt.Print("null")
	case JSONBoolean:
		fmt.Print(v.Value)
	case JSONNumber:
		fmt.Print(v.Value)
	case JSONString:
		fmt.Printf("%#v", v.Value)
	}
}
func printIndentation(indentationLevel int) {
	for i := 0; i < indentationLevel; i++ {
		fmt.Print("  ") // You can adjust the number of spaces as needed
	}
}

func (parser *Parser) parseValue() JSONValue {
	if parser.lookahead.name == "LBracket" {
		return parser.parseObject()
	} else if parser.lookahead.name == "LSquareBracket" {
		return parser.parseArray()
	} else if parser.lookahead.name == "string" {
		val := parser.match("string")
		return JSONString{
			Value: val.value,
		}
	} else if parser.lookahead.name == "number" {
		val := parser.match("number")
		float, err := strconv.ParseFloat(val.value, 64)
		if err != nil {
			return nil
		}
		return JSONNumber{
			Value: float,
		}
	} else if parser.lookahead.name == "true" || parser.lookahead.name == "false" || parser.lookahead.name == "<nil>" {
		val := parser.match(parser.lookahead.name)
		if val.value == "<nil>" {
			return JSONNull{}
		}
		parseBool, err := strconv.ParseBool(val.value)
		if err != nil {
			return nil
		}
		return JSONBoolean{
			Value: parseBool,
		}
	} else {
		log.Fatalf("Error while parsing value token: %s=%s", parser.lookahead.name, parser.lookahead.value)
	}

	return nil
}

func (parser *Parser) parseObject() JSONValue {
	parser.match("LBracket")
	obj := JSONObject{
		members: make(map[string]JSONValue),
	}
	if parser.lookahead.name == "string" {
		val := parser.match("string")
		parser.match("Colon")
		obj.members[val.value] = parser.parseValue()
		for parser.lookahead.name == "Comma" {
			parser.match("Comma")
			val := parser.match("string")
			parser.match("Colon")
			obj.members[val.value] = parser.parseValue()
		}
	}
	parser.match("RBracket")
	return obj
}

func (parser *Parser) parseArray() JSONValue {
	arr := JSONArray{elements: make([]JSONValue, 0)}
	parser.match("LSquareBracket")
	if parser.lookahead.name == "number" ||
		parser.lookahead.name == "string" ||
		parser.lookahead.name == "LSquareBracket" ||
		parser.lookahead.name == "LBracket" ||
		parser.lookahead.name == "true" ||
		parser.lookahead.name == "false" ||
		parser.lookahead.name == "<nil>" {

		arr.elements = append(arr.elements, parser.parseValue())
		for parser.lookahead.name == "Comma" {
			parser.match("Comma")
			arr.elements = append(arr.elements, parser.parseValue())
		}
	}
	parser.match("RSquareBracket")
	return arr
}
