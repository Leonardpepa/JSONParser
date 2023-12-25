# Minimal JSON Parser written in go

## Purpose
This project is a solution for [Write Your Own JSON Parser](https://codingchallenges.fyi/challenges/challenge-json-parser)
build for my personal educational purposes

# Features
* Parse json into interface{}
* TODO: implement json stringify

# Implementation Details
The implementation is based on the json specification [Introducing JSON](https://www.json.org/json-en.html).

## Lexical analysis
This step is responsible to create the tokens.
A json scanner was implemented from scratch to accomplish this task.

## Syntax analysis
This step is responsible to validate the correct structure that matches the formal grammar and create the syntax tree.
The parser implemented in this project is a recursive descent parser based on the json context free grammar seen in the specification [Introducing JSON](https://www.json.org/json-en.html).

## Syntax tree 
The json is parsed directly as an interface{}. Can be used exactly like go manipulates [Generic JSON](https://go.dev/blog/json#generic-json-with-interface)

# Example Usage
```go
package main

import (
	"JSONParser/JSONParser"
	"JSONParser/Util"
	"log"
	"os"
)

func main() {
	input, err := os.ReadFile("tests/step4/valid2.json")

	if err != nil {
		log.Fatal(err)
	}

	parsed, err := JSONParser.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	Util.Printify(parsed)
}
```
## File tests/step4/valid2.json
```json
{
  "key": "value",
  "y": "\u0020",
  "e": 12.5e-4,
  "e2": 12e+4,
  "e3": 12e-4,
  "key-n": 101000,
  "n" : -12445.1,
  "key-o": {
    "inner key": "inner value"
  },
  "key-l": [
    "list value"
  ],
  "l" : [1, 2, "dd", 3],
  "nested": {
    "n": {
      "attr": true
    }
  }
}
```
## Output
```terminal
{
  "key-n": 101000,
  "n": -12445.1,
  "key-l": ["list value"],
  "e": 0.00125,
  "e2": 120000,
  "e3": 0.0012,
  "key-o": {
    "inner key": "inner value"
  },
  "l": [1, 2, "dd", 3],
  "nested": {
    "n": {
      "attr": true
    }
  },
  "key": "value",
  "y": " "
}
```

# Tests
The parser is tested comparing the results against the native go json package.
Run the tests ```go test ./...```