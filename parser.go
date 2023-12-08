package main

import (
	"fmt"
)

type JSONParser struct {
	lexer         *JSONLexer
	lookahead     Token
	errorOccurred error
}

func nameFromType(tType string) string {
	switch tType {
	case "string":
		return "string"
	case "number":
		return "number"
	case "null", "true", "false":
		return "literal"
	case "Comma":
		return ","
	case "Colon":
		return ":"
	case "LBracket":
		return "{"
	case "RBracket":
		return "}"
	case "LSquareBracket":
		return "["
	case "RSquareBracket":
		return "]"
	default:
		return "unknown token type"
	}
}

func (parser *JSONParser) match(tType string) (Token, error) {
	var err error
	var prev Token
	var nextToken Token
	if parser.lookahead.name == tType {
		var nestedErr error
		nextToken, nestedErr = parser.lexer.getNextToken()
		if nestedErr != nil {
			err = nestedErr
		}
	} else {
		return Token{}, fmt.Errorf("type mismatch expected %s=\"%v\" got %s=\"%v\", line %d, col %d", tType, nameFromType(tType), parser.lookahead.name, parser.lookahead.value, parser.lexer.line, parser.lexer.column)
	}

	if err != nil {
		return Token{}, err
	}

	prev = parser.lookahead
	parser.lookahead = nextToken

	return prev, nil
}

func (parser *JSONParser) parse(jsonBytes []byte) (interface{}, error) {
	parser.lexer = &JSONLexer{column: 0, line: 1}
	parser.lexer.readJsonText(jsonBytes)

	nextT, err := parser.lexer.getNextToken()
	if err != nil {
		return nil, err
	}

	parser.lookahead = nextT

	parsedJson, err := parser.parseValue()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.value == "EOF" {
		return parsedJson, nil
	} else {
		return nil, fmt.Errorf("invalid token %s=\"%v\" unexpected end of json", parser.lookahead.name, parser.lookahead.value)
	}
}

func (parser *JSONParser) parseValue() (interface{}, error) {
	if parser.lookahead.name == "LBracket" {
		return parser.parseObject()
	} else if parser.lookahead.name == "LSquareBracket" {
		return parser.parseArray()
	} else if parser.lookahead.name == "string" {
		val, err := parser.match("string")
		if err != nil {
			return nil, err
		}
		return val.value, nil
	} else if parser.lookahead.name == "number" {
		val, err := parser.match("number")
		if err != nil {
			return nil, err
		}
		return val.value, nil
	} else if parser.lookahead.name == "true" || parser.lookahead.name == "false" || parser.lookahead.name == "null" {
		val, err := parser.match(parser.lookahead.name)
		if err != nil {
			return nil, err
		}
		return val.value, nil
	} else {
		return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of value", parser.lookahead.value)
	}
}

func (parser *JSONParser) parseObject() (interface{}, error) {
	_, _ = parser.match("LBracket")

	obj := make(map[string]interface{})

	if parser.lookahead.name == "string" {
		key, err := parser.match("string")
		if err != nil {
			return nil, err
		}
		_, err = parser.match("Colon")
		if err != nil {
			return nil, fmt.Errorf("invalid token \"%v\" looking for Colon=\":\"", parser.lookahead.value)
		}
		val, err := parser.parseValue()
		if err != nil {
			return nil, err
		}
		obj[key.value.(string)] = val

		for parser.lookahead.name == "Comma" {
			_, err := parser.match("Comma")
			if err != nil {
				return nil, err
			}
			key, err := parser.match("string")
			if err != nil {
				return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of object key string", parser.lookahead.value)
			}

			_, err = parser.match("Colon")
			if err != nil {
				return nil, fmt.Errorf("invalid token \"%v\" looking for Colon=\":\"", parser.lookahead.value)
			}

			val, err = parser.parseValue()
			if err != nil {
				return nil, err
			}
			obj[key.value.(string)] = val
		}
		if parser.lookahead.name != "RBracket" {
			return nil, fmt.Errorf("invalid token \"%v\" looking for a comma or an object closing }", parser.lookahead.value)
		}
	} else if parser.lookahead.name != "RBracket" {
		return nil, fmt.Errorf("invalid token \"%v\" looking object closing }", parser.lookahead.value)
	}
	_, _ = parser.match("RBracket")

	return obj, nil
}

func (parser *JSONParser) parseArray() (interface{}, error) {
	array := make([]interface{}, 0)
	_, err := parser.match("LSquareBracket")
	if err != nil {
		return nil, err
	}
	if parser.lookahead.name == "number" ||
		parser.lookahead.name == "string" ||
		parser.lookahead.name == "LSquareBracket" ||
		parser.lookahead.name == "LBracket" ||
		parser.lookahead.name == "true" ||
		parser.lookahead.name == "false" ||
		parser.lookahead.name == "null" {

		value, err := parser.parseValue()
		if err != nil {
			return nil, err
		}
		array = append(array, value)
		for parser.lookahead.name == "Comma" {
			_, err := parser.match("Comma")
			if err != nil {
				return nil, err
			}
			value, err = parser.parseValue()
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		if parser.lookahead.name != "RSquareBracket" {
			return nil, fmt.Errorf("invalid token \"%v\" looking for a comma or an ending of the array", parser.lookahead.value)
		}
	} else if parser.lookahead.name != "RSquareBracket" {
		return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of a value or an ending of the array", parser.lookahead.value)
	}
	_, _ = parser.match("RSquareBracket")

	return array, nil
}
