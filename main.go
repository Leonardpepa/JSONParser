package main

func main() {

	jsonStr := []byte(`{
  "key": "value",
  "key-n": 101,
  "key-o": {
    "yeah" : true
  },
  "key-l": ["item1", 12],
  "nullable": null,
  "j": 122.2332323232332
}`)

	jsonParser := NewJSONParser(jsonStr)
	object, err := jsonParser.parse()

	if err != nil {
		return
	}

	Printify(object)
}
