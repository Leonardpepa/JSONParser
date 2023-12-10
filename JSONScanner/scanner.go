package JSONScanner

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	LeftBracket = iota
	RightBracket
	LeftSquareBracket
	RightSquareBracket
	Comma
	Colon
	Literal
	String
	Number
	EOF
	Minus
)

const startChar = rune('\u0020')
const endChar = rune('\U0010FFFF')

const startCapitalLetter = rune('A')
const endCapitalHexLetter = rune('F')
const startLowercaseLetter = rune('a')
const endLowercaseHexLetter = rune('f')
const endLowercaseLetter = rune('z')

const whitespace1 = rune('\u0020')
const newline = rune('\u000A')
const whitespace2 = rune('\u000D')
const whitespace3 = rune('\u0009')

type Token struct {
	Type         int
	Value        interface{}
	Line, Column int
}

type JSONLexer struct {
	Runes        []rune
	Position     int
	Line, Column int
	strBuilder   strings.Builder
}

func (lexer *JSONLexer) ReadJson(jsonBytes []byte) {
	lexer.Runes = []rune(string(jsonBytes))
}

func (lexer *JSONLexer) jump(ahead int) error {
	if lexer.eof(ahead) {
		return io.EOF
	}
	lexer.Column += ahead
	lexer.Position += ahead
	return nil
}

func (lexer *JSONLexer) eof(lookahead int) bool {
	return lexer.Position+lookahead > len(lexer.Runes)-1
}

func (lexer *JSONLexer) getNextRune() (rune, error) {
	if lexer.eof(0) {
		return 0, io.EOF
	}

	lexer.Position++

	return lexer.Runes[lexer.Position-1], nil
}

func (lexer *JSONLexer) peekNextRune(lookahead int) rune {
	if lexer.eof(lookahead) {
		return 0
	}
	return lexer.Runes[lexer.Position+lookahead]
}

func (lexer *JSONLexer) isHex(lookahead int) bool {
	lookaheadRune := lexer.peekNextRune(lookahead)
	if unicode.IsDigit(lookaheadRune) {
		return true
	}
	return (lookaheadRune >= startLowercaseLetter && lookaheadRune <= endLowercaseHexLetter) ||
		(lookaheadRune >= startCapitalLetter && lookaheadRune <= endCapitalHexLetter)
}

func (lexer *JSONLexer) tokenizeDigitOnly() (string, error) {
	lookahead := 0
	var digitBuilder strings.Builder

	for {
		if unicode.IsDigit(lexer.peekNextRune(lookahead)) == false {
			break
		}
		r, err := lexer.getNextRune()
		if err != nil {
			return "", err
		}
		digitBuilder.WriteRune(r)
	}

	return digitBuilder.String(), nil
}

func (lexer *JSONLexer) tokenizeString() (*Token, error) {
	defer lexer.strBuilder.Reset()

	for {
		r, err := lexer.getNextRune()

		if err != nil {
			return nil, err
		}

		if r < startChar || r > endChar {
			break
		}

		if r == '"' {
			break
		} else if r == '\\' {
			// next character
			escaped, err := lexer.tokenizeEscapedCharacters()
			if err != nil {
				return &Token{}, err
			}
			lexer.strBuilder.WriteRune(escaped)
		} else {
			lexer.strBuilder.WriteRune(r)
		}

	}
	value := lexer.strBuilder.String()
	line := lexer.Line
	col := lexer.Column + 1
	// count the ending double quote
	lexer.Column += utf8.RuneCountInString(value) + 1
	return &Token{
		Type:   String,
		Value:  value,
		Line:   line,
		Column: col,
	}, nil

}

func (lexer *JSONLexer) tokenizeEscapedCharacters() (rune, error) {
	r, err := lexer.getNextRune()
	if err != nil {
		return 0, nil
	}
	switch r {
	case '"':
		return '"', nil
	case '\\':
		return '\\', nil
	case '/':
		return '/', nil
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'u':

		lookahead1, lookahead2, lookahead3, lookahead4 := 0, 1, 2, 3

		if lexer.isHex(lookahead1) && lexer.isHex(lookahead2) && lexer.isHex(lookahead3) && lexer.isHex(lookahead4) {
			hex1 := string(lexer.peekNextRune(lookahead1))
			hex2 := string(lexer.peekNextRune(lookahead2))
			hex3 := string(lexer.peekNextRune(lookahead3))
			hex4 := string(lexer.peekNextRune(lookahead4))

			hexValue := fmt.Sprintf("%s%s%s%s", hex1, hex2, hex3, hex4)
			unicodePoint, err := strconv.ParseUint(hexValue, 16, 32)

			if err != nil {
				return 0, err
			}
			err = lexer.jump(4)

			if err != nil {
				return 0, err
			}

			return rune(unicodePoint), nil
		}
	}
	return 0, fmt.Errorf("invalid character in string escape code")
}

func (lexer *JSONLexer) tokenizeDigits(digitAlreadyRead rune) (*Token, error) {
	defer lexer.strBuilder.Reset()

	if digitAlreadyRead == '-' {
		if lexer.peekNextRune(0) == '0' &&
			lexer.peekNextRune(1) != 'e' &&
			lexer.peekNextRune(1) != 'E' && lexer.peekNextRune(1) != '.' {
			_, err := lexer.getNextRune()
			if err != nil {
				return nil, err
			}
			lexer.Column++
			return &Token{
				Type:   Number,
				Value:  float64(-0),
				Line:   lexer.Line,
				Column: lexer.Column,
			}, nil
		}
	}

	if digitAlreadyRead == '0' {
		if lexer.peekNextRune(0) != 'e' && lexer.peekNextRune(0) != 'E' && lexer.peekNextRune(0) != '.' {
			return &Token{
				Type:   Number,
				Value:  float64(0),
				Line:   lexer.Line,
				Column: lexer.Column,
			}, nil
		}
	}

	lexer.strBuilder.WriteRune(digitAlreadyRead)
	digits, err := lexer.tokenizeDigitOnly()

	if err != nil {
		return nil, err
	}

	lexer.strBuilder.WriteString(digits)

	// real
	if lexer.peekNextRune(0) == '.' && unicode.IsDigit(lexer.peekNextRune(1)) {
		r, err := lexer.getNextRune()
		if err != nil {
			return nil, err
		}

		lexer.strBuilder.WriteRune(r)

		digits, err := lexer.tokenizeDigitOnly()
		if err != nil {
			return nil, err
		}
		lexer.strBuilder.WriteString(digits)
	}

	// exponent
	if lexer.peekNextRune(0) == 'E' || lexer.peekNextRune(0) == 'e' {
		e, err := lexer.getNextRune()
		if err != nil {
			return nil, err
		}

		lexer.strBuilder.WriteRune(e)
		lookaheadRune1 := lexer.peekNextRune(0)
		lookaheadRune2 := lexer.peekNextRune(1)

		if (lookaheadRune1 == '+' || lookaheadRune1 == '-') && unicode.IsDigit(lookaheadRune2) || unicode.IsDigit(lookaheadRune1) {
			plusOrMinus, _ := lexer.getNextRune()
			lexer.strBuilder.WriteRune(plusOrMinus)

			digits, err := lexer.tokenizeDigitOnly()

			if err != nil {
				return nil, err
			}
			lexer.strBuilder.WriteString(digits)
		} else {
			return nil, fmt.Errorf("invalid number Literal %s", lexer.strBuilder.String())
		}
	}

	value, err := strconv.ParseFloat(lexer.strBuilder.String(), 64)
	if err != nil {
		return nil, err
	}

	lexer.strBuilder.Reset()
	lexer.strBuilder.WriteString(fmt.Sprintf("%v", value))
	strValue := lexer.strBuilder.String()

	line := lexer.Line
	col := lexer.Column
	lexer.Column += utf8.RuneCountInString(strValue)

	return &Token{
		Type:   Number,
		Value:  value,
		Line:   line,
		Column: col,
	}, nil

}

func (lexer *JSONLexer) tokenizeLiterals(r rune) (*Token, error) {
	defer lexer.strBuilder.Reset()

	lookahead := 0
	for r >= startLowercaseLetter && r <= endLowercaseLetter {
		lexer.strBuilder.WriteRune(r)

		if lookahead == 4 {
			break
		}

		r = lexer.peekNextRune(lookahead)
		lookahead++
	}

	strVal := lexer.strBuilder.String()

	if strVal != "null" && strVal != "true" && strVal != "false" {
		return nil, fmt.Errorf("unrecognised Literal %s", strVal)
	}

	var value interface{}

	if strVal == "null" {
		value = nil
	} else {
		var err error
		value, err = strconv.ParseBool(strVal)
		if err != nil {
			return nil, err
		}
	}

	err := lexer.jump(len(strVal) - 1)

	if err != nil {
		return nil, err
	}

	return &Token{
		Type:   Literal,
		Value:  value,
		Line:   lexer.Line,
		Column: lexer.Column,
	}, nil
}

func (lexer *JSONLexer) GetNextToken() (*Token, error) {

	r, err := lexer.getNextRune()
	lexer.Column++

	for r == whitespace1 || r == whitespace2 || r == whitespace3 || r == newline {
		lexer.Column++
		if r == newline {
			lexer.Line++
			lexer.Column = 1
		}
		r, err = lexer.getNextRune()
	}

	if err != nil && err.Error() == "EOF" {
		return &Token{Type: EOF, Value: "EOF", Line: lexer.Line, Column: lexer.Column}, nil
	}

	if r == '{' {
		return &Token{
			Type:   LeftBracket,
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}
	if r == '}' {
		return &Token{
			Type:   RightBracket,
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == '[' {
		return &Token{
			Type:   LeftSquareBracket,
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}
	if r == ']' {
		return &Token{
			Type:   RightSquareBracket,
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == ',' {
		return &Token{
			Type:   Comma,
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == ':' {
		return &Token{
			Type:   Colon,
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == '-' {
		if unicode.IsDigit(lexer.peekNextRune(0)) {
			return lexer.tokenizeDigits(r)
		}
		return &Token{
			Type:   Minus,
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == '"' {
		return lexer.tokenizeString()
	}

	if unicode.IsDigit(r) {
		return lexer.tokenizeDigits(r)
	}

	// null, true, false
	if r == 'n' || r == 't' || r == 'f' {
		return lexer.tokenizeLiterals(r)
	}

	return nil, fmt.Errorf("unrecognised character %c=%d", r, r)
}
