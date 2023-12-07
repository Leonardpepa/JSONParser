package main

import "log"

func main() {

	input := []byte(`{
    "JSON Test Pattern pass3": {
        "The outermost value": "must be an object or array.",
        "In this test": "It is an object."
    }, "nullable": null
}`)

	parser := NewJSONParser(input)

	parsed, err := parser.parse()

	if err != nil {
		log.Fatal(err)
	}

	Printify(parsed)

}
