package main

import (
	"JSONParser/JSONParser"
	"reflect"
	"testing"
)

func TestFields(t *testing.T) {

	t.Run("Should be serialized in map[string]interface{}", func(t *testing.T) {
		input := []byte(`{
					"albumId": 2,
					"id": 51,
					"title": "non sunt voluptatem placeat consequuntur rem incidunt",
					"url": "https://via.placeholder.com/600/8e973b",
					"thumbnailUrl": "https://via.placeholder.com/150/8e973b",
					"array": [1,2,100]
				  }`)

		expected := map[string]interface{}{
			"albumId":      float64(2),
			"id":           float64(51),
			"title":        "non sunt voluptatem placeat consequuntur rem incidunt",
			"url":          "https://via.placeholder.com/600/8e973b",
			"thumbnailUrl": "https://via.placeholder.com/150/8e973b",
			"array":        []interface{}{float64(1), float64(2), float64(100)},
		}

		parsed, err := JSONParser.Parse(input)

		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		object := parsed.(map[string]interface{})
		if !reflect.DeepEqual(expected, object) {
			t.Errorf("expected %v got %v", expected, object)
		}
	})

}
