package JSONParser

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestCompareParserToNativeLib(t *testing.T) {
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

	//cases := []string{"tests/big/big.json"}

	t.Run("Comparing JSONParser with native go's JSONParser", func(t *testing.T) {
		for _, filename := range cases {
			file, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf(err.Error())
			}
			ourJson, err := Parse(file)
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

func TestCompareParserInvalidJson(t *testing.T) {
	t.Run("Parser should fail", func(t *testing.T) {
		for i := range make([]int, 33) {
			filename := fmt.Sprintf("../tests/test/fail%d.json", i+1)
			input, err := os.ReadFile(filename)
			if err != nil {
				log.Fatal(err)
			}

			_, err = Parse(input)

			if err == nil {
				t.Errorf("file %s parsed invalid json", filename)
			}
		}
	})
}
