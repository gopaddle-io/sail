package json

import (
	"github.com/yasuyuky/jsonpath"
	"io"
)

//Get String
func GetStringFromJPath(r io.Reader, path []interface{}) (string, error) {
	res, err := jsonpath.ReadString(r, path, "")
	return res, err
}

//Get Int
func GetIntFromJPath(r io.Reader, path []interface{}) (float64, error) {
	res, err := jsonpath.ReadNumber(r, path, 0)
	return res, err
}

//Get Bool
func GetBoolFromJPath(r io.Reader, path []interface{}) (bool, error) {
	res, err := jsonpath.ReadBool(r, path, false)
	return res, err
}

//Get RawJSON
func GetJSONFromJPath(r io.Reader, path []interface{}) (JSON, error) {
	res, err := jsonpath.Read(r, path, "")
	return NewFromStruct(res), err
}

//Get RawJSONArray
func GetJSONArrayFromJPath(r io.Reader, path []interface{}) ([]JSON, error) {
	res, err := jsonpath.Read(r, path, "")
	return ParseJSONArrayInterface(res), err
}
