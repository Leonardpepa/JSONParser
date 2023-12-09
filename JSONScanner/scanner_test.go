package JSONScanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestCompareScannerToNativeLib(t *testing.T) {
	cases := []string{"../tests/step1/valid.json",
		"../tests/step2/valid.json",
		"../tests/step2/valid2.json",
		"../tests/step3/valid.json",
		"../tests/step4/valid.json",
		"../tests/step4/valid2.json",
		"../tests/big/posts.json",
		"../tests/big/photos.json",
		"../tests/big/bitcoin.json",
		"../tests/big/big.json",
		"../tests/test/pass1.json",
		"../tests/test/pass2.json",
		"../tests/test/pass3.json"}

	//cases := []string{"tests/step4/valid.json"}

	t.Run("Comparing JSONScanner with native go's JSONScanner", func(t *testing.T) {
		for _, filename := range cases {

			fileRead, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf(err.Error())
			}

			jsonLexer := JSONLexer{
				Runes:  []rune{},
				Line:   1,
				Column: 0,
			}

			jsonLexer.ReadJsonText(fileRead)
			jDecoder := json.NewDecoder(bytes.NewReader(fileRead))

			for {
				token, err := jsonLexer.GetNextToken()
				if err != nil {
					t.Errorf(err.Error())
				}
				if token.Name == "Colon" || token.Name == "Comma" {
					continue
				}

				goToken, goJsonError := jDecoder.Token()

				if token.Name == "EOF" && goJsonError == io.EOF {
					break
				}

				if token.Name == "EOF" || goJsonError == io.EOF {
					t.Errorf("Lexer's didnt finish together")
				} else if goJsonError != nil {
					t.Errorf(goJsonError.Error())
				}

				goTokenValue := fmt.Sprintf("%v", goToken)
				ourTokenValue := fmt.Sprintf("%v", token.Value)

				if ourTokenValue != goTokenValue {
					t.Errorf("expected %v, got %v", goTokenValue, token.Value)
				}

			}
		}
	})
}
