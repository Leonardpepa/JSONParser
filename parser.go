package main

import (
	"fmt"
	"log"
)

type JSONParser struct {
	lexer         *JSONLexer
	lookahead     Token
	errorOccurred error
}

func NewJSONParser(jsonBytes []byte) JSONParser {
	parser := JSONParser{}
	parser.lexer = &JSONLexer{column: 0, line: 1}
	parser.lexer.readJsonText(jsonBytes)
	return parser
}

func (parser *JSONParser) match(tType string) Token {
	if parser.lookahead.name == tType {
		nextToken, err := parser.lexer.getNextToken()
		if err != nil {
			return Token{}
		}
		prev := parser.lookahead
		parser.lookahead = nextToken
		return prev
	}
	log.Fatalf("type mismatch expected %s=%v got %s=%v, line %d, col %d", tType, "", parser.lookahead.name, parser.lookahead.value, parser.lexer.line, parser.lexer.column)
	return Token{}
}

func (parser *JSONParser) parse() (interface{}, error) {
	nextT, err := parser.lexer.getNextToken()
	if err != nil {
		return nil, err
	}

	parser.lookahead = nextT

	parsedJson := parser.parseValue()

	if parser.lookahead.value == "EOF" {
		return parsedJson, nil
	} else {
		return nil, fmt.Errorf("an Error occurred")
	}
}

func (parser *JSONParser) parseValue() interface{} {
	if parser.lookahead.name == "LBracket" {
		return parser.parseObject()
	} else if parser.lookahead.name == "LSquareBracket" {
		return parser.parseArray()
	} else if parser.lookahead.name == "string" {
		val := parser.match("string")
		return val.value

	} else if parser.lookahead.name == "number" {
		val := parser.match("number")
		return val.value
	} else if parser.lookahead.name == "true" || parser.lookahead.name == "false" || parser.lookahead.name == "null" {
		val := parser.match(parser.lookahead.name)
		return val.value
	} else {
		log.Fatalf("Error while parsing value token: %s=%s", parser.lookahead.name, parser.lookahead.value)
	}

	return nil
}

func (parser *JSONParser) parseObject() interface{} {
	parser.match("LBracket")
	obj := make(map[string]interface{})

	if parser.lookahead.name == "string" {
		val := parser.match("string")
		parser.match("Colon")
		obj[val.value.(string)] = parser.parseValue()
		for parser.lookahead.name == "Comma" {
			parser.match("Comma")
			val := parser.match("string")
			parser.match("Colon")
			obj[val.value.(string)] = parser.parseValue()
		}
	}
	parser.match("RBracket")
	return obj
}

func (parser *JSONParser) parseArray() interface{} {
	arr := make([]interface{}, 0)
	parser.match("LSquareBracket")
	if parser.lookahead.name == "number" ||
		parser.lookahead.name == "string" ||
		parser.lookahead.name == "LSquareBracket" ||
		parser.lookahead.name == "LBracket" ||
		parser.lookahead.name == "true" ||
		parser.lookahead.name == "false" ||
		parser.lookahead.name == "null" {

		arr = append(arr, parser.parseValue())
		for parser.lookahead.name == "Comma" {
			parser.match("Comma")
			arr = append(arr, parser.parseValue())
		}
	}
	parser.match("RSquareBracket")
	return arr
}
