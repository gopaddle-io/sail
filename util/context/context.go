package context

import (
	"gopaddle/sail/util/json"
	"log"
	"reflect"
)

type handle struct {
	O        interface{}
	Property map[string]interface{}
}

var instantiated *handle

func Instance() *handle {
	if instantiated == nil {
		instantiated = new(handle)
		instantiated.Property = make(map[string]interface{})
	}
	return instantiated
}

func (h *handle) Get(key string) string {
	if v, ok := h.Property[key]; ok {
		return v.(string)
	}
	return ""
}

func (h *handle) Set(key string, value string) {
	h.Property[key] = value
}

func (h *handle) SetObject(key string, value interface{}) {
	h.Property[key] = value
}

func (h *handle) SetArray(key string, value []byte) {
	h.Property[key] = value
}

func (h *handle) GetObject(key string) interface{} {
	value := h.Property[key]
	log.Println("Found ", reflect.TypeOf(value))
	return value
}

func (h *handle) GetJSON(key string) json.JSON {
	if v, ok := h.Property[key]; ok {
		return v.(json.JSON)
	}
	return json.New()
}

func (h *handle) GetJSONArray(key string) []json.JSON {
	if v, ok := h.Property[key]; ok {
		return v.([]json.JSON)
	}
	var arrJSON []json.JSON
	return arrJSON
}

func (h *handle) SetJSON(key string, value json.JSON) {
	h.Property[key] = value
}
