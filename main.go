package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
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
	value        string
	line, column int
}

type JSONLexer struct {
	runes        []rune
	position     int
	line, column int
	strBuilder   strings.Builder
	token        Token
}

func (lexer *JSONLexer) jump(ahead int) {
	if lexer.EOF(ahead) == false {
		lexer.column += ahead
		lexer.position += ahead
	}
}

func (lexer *JSONLexer) EOF(lookahead int) bool {
	return lexer.position+lookahead > len(lexer.runes)-1
}

func (lexer *JSONLexer) getNextRune() (rune, error) {
	if lexer.EOF(0) {
		return 0, io.EOF
	}

	if lexer.position != 0 {
		lexer.column++
	}

	if lexer.runes[lexer.position] == Newline {
		lexer.line++
		lexer.column = 1
	}

	lexer.position++

	return lexer.runes[lexer.position-1], nil
}

func (lexer *JSONLexer) peekNextRune(lookahead int) rune {
	if lexer.EOF(0) {
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

		// TODO case '\'
		if r == DoubleQuotes {
			break
		} else if r == BackSlash {
			// next character
			escaped := lexer.tokenizeEscapedCharacters()
			builder.WriteString(escaped)
		} else {
			builder.WriteRune(r)
		}

	}
	value := builder.String()

	return Token{
		name:   "string",
		value:  value,
		line:   lexer.line,
		column: lexer.column,
	}, nil

}

func (lexer *JSONLexer) tokenizeEscapedCharacters() string {
	r, _ := lexer.getNextRune()

	switch r {
	case DoubleQuotes:
		return "\""
	case BackSlash:
		return "\\"
	case Slash:
		return "/"
	case B:
		return "\b"
	case F:
		return "\f"
	case N:
		return "\n"
	case R:
		return "\r"
	case T:
		return "\t"
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
				log.Fatal(err)
			}

			builder.WriteRune(rune(unicodePoint))
		}
		return builder.String()
	}
	return ""
}

func (lexer *JSONLexer) tokenizeDigits(digitAlreadyRead rune) (Token, error) {
	lexer.strBuilder.WriteRune(digitAlreadyRead)
	lexer.strBuilder.WriteString(lexer.tokenizeDigitOnly())

	// real
	if lexer.peekNextRune(0) == Period && unicode.IsDigit(lexer.peekNextRune(1)) {
		r, _ := lexer.getNextRune()
		lexer.strBuilder.WriteRune(r)
		lexer.strBuilder.WriteString(lexer.tokenizeDigitOnly())
		//realNumber = true
	}

	// exponent
	if lexer.peekNextRune(0) == E || lexer.peekNextRune(0) == e {
		e, _ := lexer.getNextRune()
		lookaheadRune1 := lexer.peekNextRune(0)
		lookaheadRune2 := lexer.peekNextRune(1)

		if (lookaheadRune1 == Plus || lookaheadRune1 == Minus) && unicode.IsDigit(lookaheadRune2) {
			plusOrMinus, _ := lexer.getNextRune()
			lexer.strBuilder.WriteRune(e)
			lexer.strBuilder.WriteRune(plusOrMinus)
			lexer.strBuilder.WriteString(lexer.tokenizeDigitOnly())
		}
	}

	float, err := strconv.ParseFloat(lexer.strBuilder.String(), 64)

	if err != nil {
		return Token{}, err
	}

	lexer.strBuilder.Reset()
	lexer.strBuilder.WriteString(fmt.Sprintf("%v", float))

	value := lexer.strBuilder.String()
	lexer.strBuilder.Reset()

	return Token{
		name:   "number",
		value:  value,
		line:   lexer.line,
		column: lexer.column,
	}, nil

}

func (lexer *JSONLexer) tokenizeLiterals(r rune) (Token, error, bool) {
	lookahead := 0
	var strBuilder strings.Builder
	for copyR := r; copyR >= StartLowercaseLetter && copyR <= EndLowercaseLetter; {
		strBuilder.WriteRune(copyR)
		copyR = lexer.peekNextRune(lookahead)
		value := strBuilder.String()
		if len(value) >= 3 {
			if value == Null || value == True || value == False {
				if value == Null {
					value = "<nil>"
				}
				lexer.jump(lookahead)
				return Token{
					name:   value,
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

	for unicode.IsSpace(r) {
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
			name:   "Comma",
			value:  string(r),
			line:   lexer.line,
			column: lexer.column,
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
		return lexer.tokenizeDigits(r)
	}
	// true, false, null
	token, err, done := lexer.tokenizeLiterals(r)
	if done {
		return token, err
	}

	return Token{}, fmt.Errorf("unrecognised character %c=%d", r, r)
}

func main() {
	//fmt.Printf("Lexical analysis testing using my Implementation compared to the go's json/Decoder lib")
	//testMyLexerWithNativeJsonLib("tests/step1/valid.json")
	//testMyLexerWithNativeJsonLib("tests/step2/valid.json")
	//testMyLexerWithNativeJsonLib("tests/step2/valid2.json")
	//testMyLexerWithNativeJsonLib("tests/step3/valid.json")
	//testMyLexerWithNativeJsonLib("tests/step4/valid.json")
	//testMyLexerWithNativeJsonLib("tests/step4/valid2.json")
	//testMyLexerWithNativeJsonLib("tests/big/posts.json")
	//testMyLexerWithNativeJsonLib("tests/big/photos.json")
	//testMyLexerWithNativeJsonLib("tests/big/bitcoin.json")
	//testMyLexerWithNativeJsonLib("tests/big/big.json")

	text, err := os.ReadFile("tests/big/posts.json")
	if err != nil {
		log.Fatal(err.Error())
	}

	jsonLexer := JSONLexer{
		runes:  []rune{},
		line:   1,
		column: 1,
	}

	jsonLexer.runes = []rune(string(text))

	err = jsonParser(&jsonLexer)

	if err != nil {
		log.Fatal(err.Error())
	}

}

func match(lexer *JSONLexer, tType string) bool {
	if lexer.token.name == tType {
		nextToken, err := lexer.getNextToken()
		if err != nil {
			return false
		}
		lexer.token = nextToken
		return true
	}
	return false
}

func jsonParser(lexer *JSONLexer) error {
	nextT, err := lexer.getNextToken()
	if err != nil {
		return err
	}
	lexer.token = nextT
	parseElement(lexer)
	if lexer.token.value == "EOF" {
		fmt.Println("\nParsing completed")
		return nil
	} else {
		return fmt.Errorf("an Error occurred")
	}
}

func parseValue(lexer *JSONLexer) {
	if lexer.token.name == "LBracket" {
		parseObject(lexer)
	} else if lexer.token.name == "LSquareBracket" {
		parseArray(lexer)
	} else if lexer.token.name == "string" {
		match(lexer, "string")
	} else if lexer.token.name == "number" {
		match(lexer, "number")
	} else if lexer.token.name == "true" || lexer.token.name == "false" || lexer.token.name == "<nil>" {
		match(lexer, lexer.token.name)
	} else {
		log.Fatalf("Error while parsing value token: %s=%s", lexer.token.name, lexer.token.value)
	}
}

func parseObject(lexer *JSONLexer) {
	match(lexer, "LBracket")
	if lexer.token.name == "string" {
		parseMembers(lexer)
	}
	match(lexer, "RBracket")
}

func parseMembers(lexer *JSONLexer) {
	parseMember(lexer)
	if lexer.token.name == "Comma" {
		match(lexer, "Comma")
		parseMembers(lexer)
	}
}

func parseMember(lexer *JSONLexer) {
	match(lexer, "string")
	match(lexer, "Colon")
	parseElement(lexer)
}

func parseArray(lexer *JSONLexer) {
	match(lexer, "LSquareBracket")
	if lexer.token.name == "number" ||
		lexer.token.name == "string" ||
		lexer.token.name == "LSquareBracket" ||
		lexer.token.name == "LBracket" ||
		lexer.token.name == "true" ||
		lexer.token.name == "false" ||
		lexer.token.name == "<nil>" {
		parseElements(lexer)
	}
	match(lexer, "RSquareBracket")
}

func parseElements(lexer *JSONLexer) {
	parseElement(lexer)
	if lexer.token.name == "Comma" {
		match(lexer, "Comma")
		parseElements(lexer)
	}
}

func parseElement(lexer *JSONLexer) {
	parseValue(lexer)
}

func testMyLexerWithNativeJsonLib(filename string) {
	file, err := os.Open(filename)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal()
		}
	}(file)
	if err != nil {
		log.Fatal(err.Error())
	}

	text, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err.Error())
		return
	}

	jsonLexer := JSONLexer{
		runes:  []rune{},
		line:   1,
		column: 1,
	}

	jsonLexer.runes = []rune(string(text))

	jDecoder := json.NewDecoder(file)

	var myTokens []string
	var gosTokens []string

	for {
		token, err := jsonLexer.getNextToken()
		if err != nil {
			log.Println(err.Error())
			return
		}
		if token.name == "Colon" || token.name == "Comma" {
			continue
		}
		myTokens = append(myTokens, token.value)

		goToken, err2 := jDecoder.Token()
		if err2 == io.EOF {
			gosTokens = append(gosTokens, "EOF")
		} else {
			gosTokens = append(gosTokens, fmt.Sprintf("%v", goToken))
		}

		if token.name == "EOF" && err2 == io.EOF {
			break
		} else if token.name == "EOF" || err2 == io.EOF {
			log.Println(err2.Error())
		}
	}
	fmt.Printf("Testing file: %s, tokens: %d=%d. equality: %v\n", filename, len(myTokens), len(gosTokens), slices.Equal(myTokens, gosTokens))
}
