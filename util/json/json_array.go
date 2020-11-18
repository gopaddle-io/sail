package json

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type JSONArray struct {
	array []JSON
}

/////////////////////////////////////////////////////
////////////////////// Create ///////////////////////
/////////////////////////////////////////////////////

func NewArray() (newObject JSONArray) {
	m := make([]JSON, 0)
	newObject.array = m
	return newObject
}

func NewArrayFrom(m []JSON) (newObject JSONArray) {
	newObject.array = m
	return newObject
}

/////////////////////////////////////////////////////
/////////////////////// Parse ///////////////////////
/////////////////////////////////////////////////////

func ParseJSONArrayString(data string) JSONArray {
	return ParseJSONArray([]byte(data))
}

func ParseJSONArray(data []byte) (parsed JSONArray) {
	//Array of Array not supported
	array := make([]JSON, 0)
	var f interface{}
	if err := json.Unmarshal(data, &f); err != nil {
		return
	}
	switch val := f.(type) {
	case []interface{}:
		for _, u := range val {
			switch u.(type) {
			case map[string]interface{}:
				m := u.(map[string]interface{})
				array = append(array, NewFrom(m))
			}
		}
	}
	return NewArrayFrom(array)
}

func ParseJSONArrayInterface(data interface{}) []JSON {
	//Array of Array not supported
	array := make([]JSON, 0)
	switch val := data.(type) {
	case []interface{}:
		for _, u := range val {
			switch u.(type) {
			case map[string]interface{}:
				m := u.(map[string]interface{})
				array = append(array, NewFrom(m))
			}
		}
	}
	return array
}

func ParseJSONArrayReadCloser(rc io.ReadCloser) JSONArray {
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		panic(-255)
	}
	return ParseJSONArray(data)
}

/////////////////////////////////////////////////////
////////////////////// GetValue /////////////////////
/////////////////////////////////////////////////////

func (j *JSONArray) GetAt(idx int) *JSON {
	return &j.array[idx]
}

func (j *JSONArray) Range() []JSON {
	return j.array
}

func (j *JSONArray) Size() int {
	return len(j.array)
}

/////////////////////////////////////////////////////
////////////////////// SetValue /////////////////////
/////////////////////////////////////////////////////

func (j *JSONArray) Append(jobj JSON) {
	j.array = append(j.array, jobj)
}

/////////////////////////////////////////////////////
////////////////////// ToString /////////////////////
/////////////////////////////////////////////////////

func (j *JSONArray) ToString() string {
	result := "["
	for _, jobj := range j.array {
		result += jobj.ToString() + ","
	}
	result = prune(result, ",") //prune extra comma if needed
	result += "]"
	return result
}
