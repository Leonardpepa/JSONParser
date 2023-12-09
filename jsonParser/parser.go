package jsonParser

import (
	"JSONParser/jsonScanner"
	"fmt"
)

type JSONParser struct {
	lexer         *jsonScanner.JSONLexer
	lookahead     *jsonScanner.Token
	errorOccurred error
}

func name(tType string) string {
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

func (parser *JSONParser) match(tType string) (*jsonScanner.Token, error) {
	var err error
	var prev *jsonScanner.Token
	var nextToken *jsonScanner.Token
	if parser.lookahead.Name == tType {
		var nestedErr error
		nextToken, nestedErr = parser.lexer.GetNextToken()
		if nestedErr != nil {
			err = nestedErr
		}
	} else {
		return nil, fmt.Errorf("type mismatch expected %s=\"%v\" got %s=\"%v\", Line %d, col %d", tType, name(tType), parser.lookahead.Name, parser.lookahead.Value, parser.lexer.Line, parser.lexer.Column)
	}

	if err != nil {
		return nil, err
	}

	prev = parser.lookahead
	parser.lookahead = nextToken

	return prev, nil
}

func Parse(jsonBytes []byte) (interface{}, error) {
	parser := JSONParser{}
	parser.lexer = &jsonScanner.JSONLexer{Column: 0, Line: 1}
	parser.lexer.ReadJsonText(jsonBytes)

	nextT, err := parser.lexer.GetNextToken()
	if err != nil {
		return nil, err
	}

	parser.lookahead = nextT

	parsedJson, err := parser.parseValue()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Value == "EOF" {
		return parsedJson, nil
	} else {
		return nil, fmt.Errorf("invalid token %s=\"%v\" unexpected end of json", parser.lookahead.Name, parser.lookahead.Value)
	}
}

func (parser *JSONParser) parseValue() (interface{}, error) {
	if parser.lookahead.Name == "LBracket" {
		return parser.parseObject()
	} else if parser.lookahead.Name == "LSquareBracket" {
		return parser.parseArray()
	} else if parser.lookahead.Name == "string" {
		val, err := parser.match("string")
		if err != nil {
			return nil, err
		}
		return val.Value, nil
	} else if parser.lookahead.Name == "number" {
		val, err := parser.match("number")
		if err != nil {
			return nil, err
		}
		return val.Value, nil
	} else if parser.lookahead.Name == "true" || parser.lookahead.Name == "false" || parser.lookahead.Name == "null" {
		val, err := parser.match(parser.lookahead.Name)
		if err != nil {
			return nil, err
		}
		return val.Value, nil
	} else {
		return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of Value", parser.lookahead.Value)
	}
}

func (parser *JSONParser) parseObject() (interface{}, error) {
	_, _ = parser.match("LBracket")

	obj := make(map[string]interface{})

	if parser.lookahead.Name == "string" {
		key, err := parser.match("string")
		if err != nil {
			return nil, err
		}
		_, err = parser.match("Colon")
		if err != nil {
			return nil, fmt.Errorf("invalid token \"%v\" looking for Colon=\":\"", parser.lookahead.Value)
		}
		val, err := parser.parseValue()
		if err != nil {
			return nil, err
		}
		obj[key.Value.(string)] = val

		for parser.lookahead.Name == "Comma" {
			_, err := parser.match("Comma")
			if err != nil {
				return nil, err
			}
			key, err := parser.match("string")
			if err != nil {
				return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of object key string", parser.lookahead.Value)
			}

			_, err = parser.match("Colon")
			if err != nil {
				return nil, fmt.Errorf("invalid token \"%v\" looking for Colon=\":\"", parser.lookahead.Value)
			}

			val, err = parser.parseValue()
			if err != nil {
				return nil, err
			}
			obj[key.Value.(string)] = val
		}
		if parser.lookahead.Name != "RBracket" {
			return nil, fmt.Errorf("invalid token \"%v\" looking for a comma or an object closing }", parser.lookahead.Value)
		}
	} else if parser.lookahead.Name != "RBracket" {
		return nil, fmt.Errorf("invalid token \"%v\" looking object closing }", parser.lookahead.Value)
	}
	_, err := parser.match("RBracket")
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (parser *JSONParser) parseArray() (interface{}, error) {
	array := make([]interface{}, 0)
	_, err := parser.match("LSquareBracket")
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Name == "number" ||
		parser.lookahead.Name == "string" ||
		parser.lookahead.Name == "LSquareBracket" ||
		parser.lookahead.Name == "LBracket" ||
		parser.lookahead.Name == "true" ||
		parser.lookahead.Name == "false" ||
		parser.lookahead.Name == "null" {

		Value, err := parser.parseValue()
		if err != nil {
			return nil, err
		}
		array = append(array, Value)
		for parser.lookahead.Name == "Comma" {
			_, err := parser.match("Comma")
			if err != nil {
				return nil, err
			}
			Value, err = parser.parseValue()
			if err != nil {
				return nil, err
			}
			array = append(array, Value)
		}
		if parser.lookahead.Name != "RSquareBracket" {
			return nil, fmt.Errorf("invalid token \"%v\" looking for a comma or an ending of the array", parser.lookahead.Value)
		}
	} else if parser.lookahead.Name != "RSquareBracket" {
		return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of a Value or an ending of the array", parser.lookahead.Value)
	}
	_, err = parser.match("RSquareBracket")
	if err != nil {
		return nil, err
	}
	return array, nil
}
