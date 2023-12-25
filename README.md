# Minimal JSON Parser written in go

## Purpose
This project is a solution for [Write Your Own JSON Parser](https://codingchallenges.fyi/challenges/challenge-json-parser)
build for my personal educational purposes

# Features
* Parse json into interface{}
* TODO: implement json stringify

# Implementation Details
The implementation is based in the json specification [Introducing JSON](https://www.json.org/json-en.html).

## Lexical analysis
This step is responsible to create the tokens.
A json scanner was implemented from scratch to accomplish this task.

## Syntax analysis
This step is responsible to match the tokens in the correct way that matches the formal grammar and create the syntax tree.
The parser implemented in this project is a recursive descent parser based on the json context free grammar seen in the specification [Introducing JSON](https://www.json.org/json-en.html).

## Syntax tree 
The json is parsed directly as an interface{}. Can be used exactly like go manipulates [Generic JSON](https://go.dev/blog/json#generic-json-with-interface)

# Example Usage
```go
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