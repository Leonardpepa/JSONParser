package JSONParser

import (
	"JSONParser/JSONScanner"
	"fmt"
)

type JSONParser struct {
	lexer         *JSONScanner.JSONLexer
	lookahead     *JSONScanner.Token
	errorOccurred error
}

func name(tType int) string {
	switch tType {
	case JSONScanner.String:
		return "string"
	case JSONScanner.Number:
		return "number"
	case JSONScanner.Literal:
		return "literal"
	case JSONScanner.Comma:
		return ","
	case JSONScanner.Colon:
		return ":"
	case JSONScanner.LeftBracket:
		return "{"
	case JSONScanner.RightBracket:
		return "}"
	case JSONScanner.LeftSquareBracket:
		return "["
	case JSONScanner.RightSquareBracket:
		return "]"
	default:
		return "unknown token type"
	}
}

func (parser *JSONParser) match(tType int) (*JSONScanner.Token, error) {
	var err error
	var prev *JSONScanner.Token
	var nextToken *JSONScanner.Token

	if parser.lookahead.Type == tType {
		nextToken, err = parser.lexer.GetNextToken()
	} else {
		err = fmt.Errorf("type mismatch expected %v got \"%v\", Line %d, col %d", name(tType), parser.lookahead.Value, parser.lexer.Line, parser.lexer.Column)
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
	parser.lexer = &JSONScanner.JSONLexer{Column: 0, Line: 1}
	parser.lexer.ReadJson(jsonBytes)

	nextT, err := parser.lexer.GetNextToken()
	if err != nil {
		return nil, err
	}

	parser.lookahead = nextT

	parsedJson, err := parser.parseValue()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == JSONScanner.EOF {
		return parsedJson, nil
	} else {
		return nil, fmt.Errorf("invalid token \"%v\" unexpected end of json", parser.lookahead.Value)
	}
}

func (parser *JSONParser) parseValue() (interface{}, error) {
	if parser.lookahead.Type == JSONScanner.LeftBracket {
		return parser.parseObject()
	} else if parser.lookahead.Type == JSONScanner.LeftSquareBracket {
		return parser.parseArray()
	} else if parser.lookahead.Type == JSONScanner.String {
		val, err := parser.match(JSONScanner.String)
		if err != nil {
			return nil, err
		}
		return val.Value, nil
	} else if parser.lookahead.Type == JSONScanner.Number {
		val, err := parser.match(JSONScanner.Number)
		if err != nil {
			return nil, err
		}
		return val.Value, nil
	} else if parser.lookahead.Type == JSONScanner.Literal {
		val, err := parser.match(parser.lookahead.Type)
		if err != nil {
			return nil, err
		}
		return val.Value, nil
	} else {
		return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of Value", parser.lookahead.Value)
	}
}

func (parser *JSONParser) parseObject() (interface{}, error) {
	_, err := parser.match(JSONScanner.LeftBracket)
	if err != nil {
		return nil, fmt.Errorf("%s\nlooking for beginning of object key string\n", err.Error())
	}
	obj := make(map[string]interface{})

	if parser.lookahead.Type == JSONScanner.String {
		key, err := parser.match(JSONScanner.String)
		if err != nil {
			return nil, err
		}
		_, err = parser.match(JSONScanner.Colon)
		if err != nil {
			return nil, fmt.Errorf("invalid token \"%v\" looking for Colon=\":\"", parser.lookahead.Value)
		}
		val, err := parser.parseValue()
		if err != nil {
			return nil, err
		}
		obj[key.Value.(string)] = val

		for parser.lookahead.Type == JSONScanner.Comma {
			_, err := parser.match(JSONScanner.Comma)
			if err != nil {
				return nil, fmt.Errorf("%s\nlooking for beginning of object key string\n", err.Error())
			}
			key, err := parser.match(JSONScanner.String)
			if err != nil {
				return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of object key string", parser.lookahead.Value)
			}

			_, err = parser.match(JSONScanner.Colon)
			if err != nil {
				return nil, fmt.Errorf("invalid token \"%v\" looking for Colon=\":\"", parser.lookahead.Value)
			}

			val, err = parser.parseValue()
			if err != nil {
				return nil, err
			}
			obj[key.Value.(string)] = val
		}
		if parser.lookahead.Type != JSONScanner.RightBracket {
			return nil, fmt.Errorf("invalid token \"%v\" looking for a comma or an object closing }", parser.lookahead.Value)
		}
	} else if parser.lookahead.Type != JSONScanner.RightBracket {
		return nil, fmt.Errorf("invalid token \"%v\" looking object closing }", parser.lookahead.Value)
	}
	_, err = parser.match(JSONScanner.RightBracket)
	if err != nil {
		return nil, fmt.Errorf("%s\nlooking object closing }", err.Error())
	}
	return obj, nil
}

func (parser *JSONParser) parseArray() (interface{}, error) {
	array := make([]interface{}, 0)
	_, err := parser.match(JSONScanner.LeftSquareBracket)
	if err != nil {
		return nil, fmt.Errorf("%s\nlooking for a value or an ending of the array\n", err.Error())
	}
	if parser.lookahead.Type == JSONScanner.Number ||
		parser.lookahead.Type == JSONScanner.String ||
		parser.lookahead.Type == JSONScanner.LeftSquareBracket ||
		parser.lookahead.Type == JSONScanner.LeftBracket ||
		parser.lookahead.Type == JSONScanner.Literal {

		Value, err := parser.parseValue()
		if err != nil {
			return nil, err
		}
		array = append(array, Value)
		for parser.lookahead.Type == JSONScanner.Comma {
			_, err := parser.match(JSONScanner.Comma)
			if err != nil {
				return nil, err
			}
			Value, err = parser.parseValue()
			if err != nil {
				return nil, err
			}
			array = append(array, Value)
		}
		if parser.lookahead.Type != JSONScanner.RightSquareBracket {
			return nil, fmt.Errorf("invalid token \"%v\" looking for a comma or an ending of the array", parser.lookahead.Value)
		}
	} else if parser.lookahead.Type != JSONScanner.RightSquareBracket {
		return nil, fmt.Errorf("invalid token \"%v\" looking for beginning of a Value or an ending of the array", parser.lookahead.Value)
	}
	_, err = parser.match(JSONScanner.RightSquareBracket)
	if err != nil {
		return nil, err
	}
	return array, nil
}
