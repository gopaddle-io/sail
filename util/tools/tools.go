package tools

import (
	gson "encoding/json"
)

// CopyStruct Make a deep copy from src into dst.
// dst should be pointer vairable
func CopyStruct(src interface{}, dst interface{}) error {
	bytes, err := gson.Marshal(src)
	if err != nil {
		return err
	}
	err = gson.Unmarshal(bytes, dst)
	if err != nil {
		return err
	}
	return nil
}
