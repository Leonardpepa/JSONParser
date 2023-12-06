package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestCompareParserToNativeLib(t *testing.T) {
	cases := []string{"tests/step1/valid.json", "tests/step2/valid.json",
		"tests/step2/valid2.json",
		"tests/step3/valid.json",
		"tests/step4/valid.json",
		"tests/step4/valid2.json",
		"tests/big/posts.json",
		"tests/big/photos.json",
		"tests/big/bitcoin.json",
		"tests/big/big.json"}

	//cases := []string{"tests/step4/valid.json"}

	t.Run("Comparing scanner wit native go's scanner", func(t *testing.T) {
		for _, filename := range cases {
			file, err := os.ReadFile(filename)
			if err != nil {
				return
			}
			parser := NewJSONParser(file)

			ourJson, err := parser.parse()
			if err != nil {
				t.Errorf(err.Error())
			}

			var goJson interface{}
			err = json.Unmarshal(file, &goJson)
			if err != nil {
				t.Errorf("Go's json threw an error, %s", err.Error())
			}

			switch goJsonV := goJson.(type) {
			case map[string]interface{}:
				if reflect.DeepEqual(ourJson.(map[string]interface{}), goJsonV) == false {
					t.Errorf("mismatch expected %v got %v", goJson, ourJson)
				}
			case []interface{}:
				if reflect.DeepEqual(ourJson.([]interface{}), goJsonV) == false {
					t.Errorf("mismatch expected %v got %v", goJson, ourJson)
				}
			case string:
				if reflect.DeepEqual(ourJson.(string), goJsonV) == false {
					t.Errorf("mismatch expected %v got %v", goJson, ourJson)
				}
			case float64:
				if reflect.DeepEqual(ourJson.(float64), goJsonV) == false {
					t.Errorf("mismatch expected %v got %v", goJson, ourJson)
				}
			case bool:
				if reflect.DeepEqual(ourJson.(bool), goJsonV) == false {
					t.Errorf("mismatch expected %v got %v", goJson, ourJson)
				}
			case nil:
				if ourJson != nil {
					t.Errorf("mismatch expected %v got %v", goJson, ourJson)
				}

			}

		}
	})
}
