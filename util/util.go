package util

import (
	"fmt"
	"gopaddle/sail/util/json"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func RandomSequence(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// LoadConfig to load configs
func LoadConfig(fileName string, key string) (json.JSON, error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("Configuraiton '%s' file not found: %v", fileName, err.Error())
		return json.New(), fmt.Errorf("Configuraiton '%s' file not found: %v", fileName, err.Error())
	}
	config := json.Parse(file)
	if key == "" {
		return config, nil
	}
	return config.GetJSON(key), nil
}

func GetPtrInt64(x int64) *int64 {
	return &x
}

func GetPtrString(x string) *string {
	return &x
}

func GetPtrInt32(x int32) *int32 {
	return &x
}

func GetPtrInt(x int) *int {
	return &x
}

func GetPtrBool(b bool) *bool {
	return &b
}

func NewRequestID() string {
	return uuid.New().String()
}

// StringContains Check if exists in array
// Returns true if exist otherwise false
func StringContains(val string, array []string) bool {
	for _, v := range array {
		if val == v {
			return true
		}
	}
	return false
}

// IntContains Check if exists in array
func IntContains(val int, array []int) bool {
	for _, v := range array {
		if val == v {
			return true
		}
	}
	return false
}

func IsItInt(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return false
	}
	return true
}

func PrintStruct(v interface{}) {
	fmt.Printf("\n%+v\n\n", v)
}
