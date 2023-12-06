package main

func main() {

	jsonStr := []byte(`{"e": 0.123456789e-12,
        "E": 1.234567890E+34,
        "":  23456789012E66,
        "zero": 0,
        "one": 1}`)

	jsonParser := NewJSONParser(jsonStr)
	object, err := jsonParser.parse()

	if err != nil {
		return
	}

	Printify(object)
}
