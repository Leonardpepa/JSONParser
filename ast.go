package main

type JSONValue interface {
	getValue() interface{}
}

type JSONObject struct {
	members map[string]JSONValue
}

type JSONArray struct {
	elements []JSONValue
}

type JSONString struct {
	Value string
}

type JSONNumber struct {
	Value float64
}

type JSONBoolean struct {
	Value bool
}

type JSONNull struct {
}

func (obj JSONBoolean) getValue() interface{} {
	return obj.Value
}

func (obj JSONNumber) getValue() interface{} {
	return obj.Value
}

func (obj JSONString) getValue() interface{} {
	return obj.Value
}

func (obj JSONArray) getValue() interface{} {
	return obj.elements
}

func (obj JSONObject) getValue() interface{} {
	return obj.members
}

func (obj JSONNull) getValue() interface{} {
	return nil
}

//	func (obj JSONObject) getValue(key string) interface{} {
//		return obj.members[key]
//	}
