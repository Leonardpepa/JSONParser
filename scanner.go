package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const True = "true"
const False = "false"
const Null = "null"

const LBracket = rune('{')
const RBracket = rune('}')
const Comma = rune(',')
const Colon = rune(':')
const LSquareBracket = rune('[')
const RSquareBracket = rune(']')
const StartChar = rune('\u0020')
const EndChar = rune('\U0010FFFF')
const DoubleQuotes = rune('"')
const BackSlash = rune('\\')
const Slash = rune('/')
const B = rune('b')
const F = rune('f')
const N = rune('n')
const R = rune('r')
const T = rune('t')
const U = rune('u')
const StartCapitalLetter = rune('A')
const EndCapitalLetter = rune('Z')
const EndCapitalHexLetter = rune('F')
const StartLowercaseLetter = rune('a')
const EndLowercaseHexLetter = rune('f')
const EndLowercaseLetter = rune('z')
const StartNonZeroDigit = rune('1')
const EndNonZeroDigit = rune('9')
const Zero = rune('0')
const Period = rune('.')
const E = rune('E')
const e = rune('e')
const Plus = rune('+')
const Minus = rune('-')
const Whitespace1 = rune('\u0020')
const Newline = rune('\u000A')
const Whitespace3 = rune('\u000D')
const Whitespace4 = rune('\u0009')

type Token struct {
	name         string
	value        interface{}
	line, column int
}

type JSONLexer struct {
	runes        []rune
	position     int
	line, column int
}

func (lexer *JSONLexer) jump(ahead int) error {
	if lexer.EOF(ahead) {
		return fmt.Errorf("EOF")
	}
	lexer.column += ahead
	lexer.position += ahead
	return nil
}

func (lexer *JSONLexer) EOF(lookahead int) bool {
	return lexer.position+lookahead > len(lexer.runes)-1
}

func (lexer *JSONLexer) getNextRune() (rune, error) {
	if lexer.EOF(0) {
		return 0, io.EOF
	}

	lexer.position++

	return lexer.runes[lexer.position-1], nil
}

func (lexer *JSONLexer) peekNextRune(lookahead int) rune {
	if lexer.EOF(lookahead) {
		return 0
	}
	return lexer.runes[lexer.position+lookahead]
}

func (lexer *JSONLexer) isHex(lookahead int) bool {
	lookaheadRune := lexer.peekNextRune(lookahead)
	if unicode.IsDigit(lookaheadRune) {
		return true
	} else if (lookaheadRune >= StartLowercaseLetter && lookaheadRune <= EndLowercaseHexLetter) ||
		(lookaheadRune >= StartCapitalLetter && lookaheadRune <= EndCapitalHexLetter) {
		return true
	}
	return false
}

func (lexer *JSONLexer) tokenizeHex() string {
	var hexBuilder strings.Builder
	lookahead := 0

	if lexer.isHex(lookahead) {
		r, _ := lexer.getNextRune()
		hexBuilder.WriteRune(r)
	}

	return hexBuilder.String()
}

func (lexer *JSONLexer) tokenizeDigitOnly() string {
	lookahead := 0
	var digitBuilder strings.Builder
	for {
		if unicode.IsDigit(lexer.peekNextRune(lookahead)) == false {
			break
		}
		r, _ := lexer.getNextRune()
		digitBuilder.WriteRune(r)
	}

	return digitBuilder.String()
}

func (lexer *JSONLexer) tokenizeString() (Token, error) {
	var builder strings.Builder

	for {
		r, err := lexer.getNextRune()

		if err != nil {
			return Token{}, err
		}

		if r < StartChar || r > EndChar {
			break
		}

		if r == DoubleQuotes {
			break
		} else if r == BackSlash {
			// next character
			escaped, err := lexer.tokenizeEscapedCharacters()
			if err != nil {
				return Token{}, err
			}
			builder.WriteString(escaped)
		} else {
			builder.WriteRune(r)
		}

	}
	value := builder.String()
	line := lexer.line
	col := lexer.column
	lexer.column += utf8.RuneCountInString(value)
	return Token{
		name:   "string",
		value:  value,
		line:   line,
		column: col,
	}, nil

}

func (lexer *JSONLexer) tokenizeEscapedCharacters() (string, error) {
	r, _ := lexer.getNextRune()

	switch r {
	case DoubleQuotes:
		return "\"", nil
	case BackSlash:
		return "\\", nil
	case Slash:
		return "/", nil
	case B:
		return "\b", nil
	case F:
		return "\f", nil
	case N:
		return "\n", nil
	case R:
		return "\r", nil
	case T:
		return "\t", nil
	case U:
		var builder strings.Builder
		lookahead1, lookahead2, lookahead3, lookahead4 := 0, 1, 2, 3

		if lexer.isHex(lookahead1) && lexer.isHex(lookahead2) && lexer.isHex(lookahead3) && lexer.isHex(lookahead4) {
			hex1 := lexer.tokenizeHex()
			hex2 := lexer.tokenizeHex()
			hex3 := lexer.tokenizeHex()
			hex4 := lexer.tokenizeHex()

			hexValue := fmt.Sprintf("%s%s%s%s", hex1, hex2, hex3, hex4)
			unicodePoint, err := strconv.ParseUint(hexValue, 16, 32)

			if err != nil {
				return "", err
			}

			builder.WriteRune(rune(unicodePoint))
		}
		return builder.String(), nil
	default:
		return "", fmt.Errorf("invalid character in string escape code")
	}
}

func (lexer *JSONLexer) tokenizeDigits(digitAlreadyRead rune) (Token, error) {
	var strBuilder strings.Builder
	strBuilder.WriteRune(digitAlreadyRead)
	strBuilder.WriteString(lexer.tokenizeDigitOnly())

	// real
	if lexer.peekNextRune(0) == Period && unicode.IsDigit(lexer.peekNextRune(1)) {
		r, _ := lexer.getNextRune()
		strBuilder.WriteRune(r)
		strBuilder.WriteString(lexer.tokenizeDigitOnly())
		//realNumber = true
	}

	// exponent
	if lexer.peekNextRune(0) == E || lexer.peekNextRune(0) == e {
		e, _ := lexer.getNextRune()
		strBuilder.WriteRune(e)
		lookaheadRune1 := lexer.peekNextRune(0)
		lookaheadRune2 := lexer.peekNextRune(1)

		if (lookaheadRune1 == Plus || lookaheadRune1 == Minus) && unicode.IsDigit(lookaheadRune2) || unicode.IsDigit(lookaheadRune1) {
			plusOrMinus, _ := lexer.getNextRune()
			strBuilder.WriteRune(plusOrMinus)
			strBuilder.WriteString(lexer.tokenizeDigitOnly())
		} else {
			return Token{}, fmt.Errorf("invalid number literal %s", strBuilder.String())
		}
	}

	var value interface{}

	var err error
	value, err = strconv.ParseFloat(strBuilder.String(), 64)
	if err != nil {
		return Token{}, err
	}

	strBuilder.Reset()
	strBuilder.WriteString(fmt.Sprintf("%v", value))
	strValue := strBuilder.String()

	line := lexer.line
	col := lexer.column
	lexer.column += utf8.RuneCountInString(strValue)

	return Token{
		name:   "number",
		value:  value,
		line:   line,
		column: col,
	}, nil

}

func (lexer *JSONLexer) tokenizeLiterals(r rune) (Token, error, bool) {
	lookahead := 0
	var strBuilder strings.Builder
	for r >= StartLowercaseLetter && r <= EndLowercaseLetter {
		strBuilder.WriteRune(r)
		r = lexer.peekNextRune(lookahead)
		strVal := strBuilder.String()
		var value interface{}

		if len(strVal) >= 3 {
			if strVal == Null || strVal == True || strVal == False {
				if strVal == Null {
					value = nil
				} else {
					var err error
					value, err = strconv.ParseBool(strVal)
					if err != nil {
						return Token{}, err, false
					}
				}
				err := lexer.jump(lookahead)
				if err != nil {
					return Token{}, err, false
				}

				lexer.column += utf8.RuneCountInString(strVal)
				return Token{
					name:   strVal,
					value:  value,
					line:   lexer.line,
					column: lexer.column,
				}, nil, true
			}
		}
		lookahead++
	}
	return Token{}, nil, false
}

func (lexer *JSONLexer) getNextToken() (Token, error) {

	r, err := lexer.getNextRune()
	lexer.column++

	for unicode.IsSpace(r) {
		lexer.column++
		if r == '\n' {
			lexer.line++
			lexer.column = 1
		}
		r, err = lexer.getNextRune()
	}

	if err != nil && err.Error() == "EOF" {
		return Token{name: "EOF", value: "EOF", line: lexer.line, column: lexer.column}, nil
	}

	if r == LBracket {
		return Token{
			name:   "LBracket",
			value:  string(r),
			line:   lexer.line,
			column: lexer.column,
		}, nil
	}
	if r == RBracket {
		return Token{
			name:   "RBracket",
			value:  string(r),
			line:   lexer.line,
			column: lexer.column,
		}, nil
	}

	if r == LSquareBracket {
		return Token{
			name:   "LSquareBracket",
			value:  string(r),
			line:   lexer.line,
			column: lexer.column,
		}, nil
	}
	if r == RSquareBracket {
		return Token{
			name:   "RSquareBracket",
			value:  string(r),
			line:   lexer.line,
			column: lexer.column,
		}, nil
	}

	if r == Comma {
		return Token{
			name:  "Comma",
			value: string(r),
		}, nil
	}

	if r == Colon {
		return Token{
			name:   "Colon",
			value:  string(r),
			line:   lexer.line,
			column: lexer.column,
		}, nil
	}

	if r == Minus {
		if lexer.peekNextRune(0) == '0' && lexer.peekNextRune(1) != e && lexer.peekNextRune(1) != E && lexer.peekNextRune(1) != Period {
			_, err := lexer.getNextRune()
			if err != nil {
				return Token{}, err
			}
			lexer.column++
			return Token{
				name:   "number",
				value:  float64(-0),
				line:   lexer.line,
				column: lexer.column,
			}, nil
		}

		if unicode.IsDigit(lexer.peekNextRune(0)) {
			return lexer.tokenizeDigits(r)
		}
		return Token{
			name:   "Minus",
			value:  string(r),
			line:   lexer.line,
			column: lexer.column,
		}, nil
	}

	if r == DoubleQuotes {
		return lexer.tokenizeString()
	}

	if unicode.IsDigit(r) {
		if r == '0' && lexer.peekNextRune(0) != e && lexer.peekNextRune(0) != E && lexer.peekNextRune(0) != Period {
			return Token{
				name:   "number",
				value:  float64(0),
				line:   lexer.line,
				column: lexer.column,
			}, nil
		}
		return lexer.tokenizeDigits(r)
	}
	// parse true, false, null
	token, err, done := lexer.tokenizeLiterals(r)
	if done {
		return token, err
	}

	return Token{}, fmt.Errorf("unrecognised character %c=%d", r, r)
}

func (lexer *JSONLexer) readJsonFile(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lexer.runes = []rune(string(file))
	return nil
}
func (lexer *JSONLexer) readJsonText(jsonBytes []byte) {
	lexer.runes = []rune(string(jsonBytes))
}
