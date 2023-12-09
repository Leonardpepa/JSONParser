package JSONScanner

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
	Name         string
	Value        interface{}
	Line, Column int
}

type JSONLexer struct {
	Runes        []rune
	Position     int
	Line, Column int
}

func (lexer *JSONLexer) jump(ahead int) error {
	if lexer.EOF(ahead) {
		return io.EOF
	}
	lexer.Column += ahead
	lexer.Position += ahead
	return nil
}

func (lexer *JSONLexer) EOF(lookahead int) bool {
	return lexer.Position+lookahead > len(lexer.Runes)-1
}

func (lexer *JSONLexer) getNextRune() (rune, error) {
	if lexer.EOF(0) {
		return 0, io.EOF
	}

	lexer.Position++

	return lexer.Runes[lexer.Position-1], nil
}

func (lexer *JSONLexer) peekNextRune(lookahead int) rune {
	if lexer.EOF(lookahead) {
		return 0
	}
	return lexer.Runes[lexer.Position+lookahead]
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
	var builder strings.Builder

	for {
		r, err := lexer.getNextRune()

		if err != nil {
			return nil, err
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
				return &Token{}, err
			}
			builder.WriteString(escaped)
		} else {
			builder.WriteRune(r)
		}

	}
	value := builder.String()
	line := lexer.Line
	col := lexer.Column
	lexer.Column += utf8.RuneCountInString(value)
	return &Token{
		Name:   "string",
		Value:  value,
		Line:   line,
		Column: col,
	}, nil

}

func (lexer *JSONLexer) tokenizeEscapedCharacters() (string, error) {
	r, err := lexer.getNextRune()
	if err != nil {
		return "", nil
	}
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

func (lexer *JSONLexer) tokenizeDigits(digitAlreadyRead rune) (*Token, error) {
	var strBuilder strings.Builder
	strBuilder.WriteRune(digitAlreadyRead)
	digits, err := lexer.tokenizeDigitOnly()

	if err != nil {
		return nil, err
	}
	strBuilder.WriteString(digits)

	// real
	if lexer.peekNextRune(0) == Period && unicode.IsDigit(lexer.peekNextRune(1)) {
		r, err := lexer.getNextRune()
		if err != nil {
			return nil, err
		}

		strBuilder.WriteRune(r)

		digits, err := lexer.tokenizeDigitOnly()
		if err != nil {
			return nil, err
		}
		strBuilder.WriteString(digits)
	}

	// exponent
	if lexer.peekNextRune(0) == E || lexer.peekNextRune(0) == e {
		e, err := lexer.getNextRune()
		if err != nil {
			return nil, err
		}

		strBuilder.WriteRune(e)
		lookaheadRune1 := lexer.peekNextRune(0)
		lookaheadRune2 := lexer.peekNextRune(1)

		if (lookaheadRune1 == Plus || lookaheadRune1 == Minus) && unicode.IsDigit(lookaheadRune2) || unicode.IsDigit(lookaheadRune1) {
			plusOrMinus, _ := lexer.getNextRune()
			strBuilder.WriteRune(plusOrMinus)

			digits, err := lexer.tokenizeDigitOnly()

			if err != nil {
				return nil, err
			}
			strBuilder.WriteString(digits)
		} else {
			return nil, fmt.Errorf("invalid number literal %s", strBuilder.String())
		}
	}

	value, err := strconv.ParseFloat(strBuilder.String(), 64)
	if err != nil {
		return nil, err
	}

	strBuilder.Reset()
	strBuilder.WriteString(fmt.Sprintf("%v", value))
	strValue := strBuilder.String()

	line := lexer.Line
	col := lexer.Column
	lexer.Column += utf8.RuneCountInString(strValue)

	return &Token{
		Name:   "number",
		Value:  value,
		Line:   line,
		Column: col,
	}, nil

}

func (lexer *JSONLexer) tokenizeLiterals(r rune) (*Token, error, bool) {
	lookahead := 0
	var strBuilder strings.Builder
	for r >= StartLowercaseLetter && r <= EndLowercaseLetter {
		strBuilder.WriteRune(r)
		r = lexer.peekNextRune(lookahead)
		strVal := strBuilder.String()
		var value interface{}

		if len(strVal) > 3 {
			if strVal == Null || strVal == True || strVal == False {
				if strVal == Null {
					value = nil
				} else {
					var err error
					value, err = strconv.ParseBool(strVal)
					if err != nil {
						return nil, err, false
					}
				}
				err := lexer.jump(lookahead)
				if err != nil {
					return nil, err, false
				}

				lexer.Column += utf8.RuneCountInString(strVal)
				return &Token{
					Name:   strVal,
					Value:  value,
					Line:   lexer.Line,
					Column: lexer.Column,
				}, nil, true
			}
		}
		lookahead++
	}
	return nil, nil, false
}

func (lexer *JSONLexer) GetNextToken() (*Token, error) {

	r, err := lexer.getNextRune()
	lexer.Column++

	for unicode.IsSpace(r) {
		lexer.Column++
		if r == '\n' {
			lexer.Line++
			lexer.Column = 1
		}
		r, err = lexer.getNextRune()
	}

	if err != nil && err.Error() == "EOF" {
		return &Token{Name: "EOF", Value: "EOF", Line: lexer.Line, Column: lexer.Column}, nil
	}

	if r == LBracket {
		return &Token{
			Name:   "LBracket",
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}
	if r == RBracket {
		return &Token{
			Name:   "RBracket",
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == LSquareBracket {
		return &Token{
			Name:   "LSquareBracket",
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}
	if r == RSquareBracket {
		return &Token{
			Name:   "RSquareBracket",
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == Comma {
		return &Token{
			Name:  "Comma",
			Value: string(r),
		}, nil
	}

	if r == Colon {
		return &Token{
			Name:   "Colon",
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == Minus {
		if lexer.peekNextRune(0) == '0' && lexer.peekNextRune(1) != e && lexer.peekNextRune(1) != E && lexer.peekNextRune(1) != Period {
			_, err := lexer.getNextRune()
			if err != nil {
				return nil, err
			}
			lexer.Column++
			return &Token{
				Name:   "number",
				Value:  float64(-0),
				Line:   lexer.Line,
				Column: lexer.Column,
			}, nil
		}

		if unicode.IsDigit(lexer.peekNextRune(0)) {
			return lexer.tokenizeDigits(r)
		}
		return &Token{
			Name:   "Minus",
			Value:  string(r),
			Line:   lexer.Line,
			Column: lexer.Column,
		}, nil
	}

	if r == DoubleQuotes {
		return lexer.tokenizeString()
	}

	if unicode.IsDigit(r) {
		if r == '0' && lexer.peekNextRune(0) != e && lexer.peekNextRune(0) != E && lexer.peekNextRune(0) != Period {
			return &Token{
				Name:   "number",
				Value:  float64(0),
				Line:   lexer.Line,
				Column: lexer.Column,
			}, nil
		}
		return lexer.tokenizeDigits(r)
	}
	// parse true, false, null
	token, err, done := lexer.tokenizeLiterals(r)
	if done {
		return token, err
	}

	return nil, fmt.Errorf("unrecognised character %c=%d", r, r)
}

func (lexer *JSONLexer) readJsonFile(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lexer.Runes = []rune(string(file))
	return nil
}
func (lexer *JSONLexer) ReadJsonText(jsonBytes []byte) {
	lexer.Runes = []rune(string(jsonBytes))
}
