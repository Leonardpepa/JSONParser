package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestCompareScannerToNativeLib(t *testing.T) {
	cases := []string{"tests/step1/valid.json", "tests/step2/valid.json",
		"tests/step2/valid2.json",
		"tests/step3/valid.json",
		"tests/step4/valid.json",
		"tests/step4/valid2.json",
		"tests/big/posts.json",
		"tests/big/photos.json",
		"tests/big/bitcoin.json",
		"tests/big/big.json"}

	t.Run("Comparing scanner wit native go's scanner", func(t *testing.T) {
		for _, filename := range cases {

			fileRead, err := os.ReadFile(filename)
			if err != nil {
				return
			}

			jsonLexer := JSONLexer{
				runes:  []rune{},
				line:   1,
				column: 0,
			}

			jsonLexer.readJsonText(fileRead)
			jDecoder := json.NewDecoder(bytes.NewReader(fileRead))

			for {
				token, err := jsonLexer.getNextToken()
				if err != nil {
					t.Errorf(err.Error())
				}
				if token.name == "Colon" || token.name == "Comma" {
					continue
				}

				goToken, goJsonError := jDecoder.Token()

				if token.name == "EOF" && goJsonError == io.EOF {
					break
				}

				if token.name == "EOF" || goJsonError == io.EOF {
					t.Errorf("EOF Error")
				} else if goJsonError != nil {
					t.Errorf(goJsonError.Error())
				}

				goTokenValue := fmt.Sprintf("%v", goToken)
				if token.value != goTokenValue {
					t.Errorf("expected %s, got %s", goTokenValue, token.value)
				}

			}
		}
	})
}
