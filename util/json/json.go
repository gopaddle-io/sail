package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type JSON struct {
	jmap map[string]interface{}
}

/////////////////////////////////////////////////////
////////////////////// Create ///////////////////////
/////////////////////////////////////////////////////

func New() (newObject JSON) {
	m := make(map[string]interface{})
	newObject.jmap = m
	return newObject
}

func NewFrom(m map[string]interface{}) (newObject JSON) {
	newObject.jmap = m
	return newObject
}

//This is ment for small/tiny json object only

func NewFromStruct(s interface{}) JSON {
	json_string := ToJSONString(s)
	return ParseString(json_string)
}

/////////////////////////////////////////////////////
/////////////////////// Parse ///////////////////////
/////////////////////////////////////////////////////

func ParseString(data string) JSON {
	return Parse([]byte(data))
}

func Parse(data []byte) (parsed JSON) {
	//Currently it supports top level map not array
	var f interface{}
	if err := json.Unmarshal(data, &f); err != nil {
		return
	}
	switch f.(type) {
	case []interface{}:
		//	log.Println("Found Array, Not Supported")
		return
	default:
		m := f.(map[string]interface{})
		//fmt.Printf("Pased Value of m: %v", m["test"])
		//str := fmt.Sprintf("%v", reflect.TypeOf(m["test"]))
		//fmt.Println("Value of str: ", str)
		parsed.jmap = m
		//fmt.Printf("Pased Value of m: %v", parsed.jmap)
		return parsed
	}
}

func ParseReadCloser(rc io.ReadCloser) JSON {
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		panic(-255)
	}
	return Parse(data)
}

/////////////////////////////////////////////////////
////////////////////// Availble /////////////////////
/////////////////////////////////////////////////////

func (jobj *JSON) HasKey(k string) bool {
	if _, ok := jobj.jmap[k]; ok {
		return true
	}
	return false
}

/////////////////////////////////////////////////////
//////////////////////GetKeyList/////////////////////
/////////////////////////////////////////////////////

func (jobj *JSON) GetKeyList() []string {
	keys := make([]string, 0)
	for k, _ := range jobj.jmap {
		keys = append(keys, k)
	}
	return keys
}

/////////////////////////////////////////////////////
////////////////////// SetValue /////////////////////
/////////////////////////////////////////////////////

func (jobj *JSON) Put(k string, v interface{}) {
	jobj.jmap[k] = v
}

/////////////////////////////////////////////////////
////////////////////// GetValue /////////////////////
/////////////////////////////////////////////////////

func (jobj *JSON) Get(k string) interface{} {
	return jobj.jmap[k]
}

func (jobj *JSON) GetBool(k string) bool {
	data := jobj.jmap[k]
	if bdata, ok := data.(bool); ok {
		return bdata
	}
	return false
}

func (jobj *JSON) GetString(k string) string {
	data := jobj.jmap[k]
	if str, ok := data.(string); ok {
		return str
	}
	return ""
}

func (jobj *JSON) GetInt(k string) int {
	data := jobj.jmap[k]
	if idata, ok := data.(int); ok {
		return idata
	} else if idata, ok := data.(float64); ok {
		return int(idata)
	}
	return 0
}

func (j *JSON) GetJSON(k string) JSON {
	f := j.jmap[k]
	switch f.(type) {
	case string:
		return New()

	case []interface{}:
		return New()

	case interface{}:
		m := f.(map[string]interface{})
		return NewFrom(m)

	default:
		return New()

	}
}

func (j *JSON) GetJSONArray(k string) []JSON {
	f := j.jmap[k]
	jarr := make([]JSON, 0)
	switch val := f.(type) {
	case []interface{}:
		for _, u := range val {
			switch u.(type) {
			case map[string]interface{}:
				m := u.(map[string]interface{})
				jarr = append(jarr, NewFrom(m))
			}
		}
	}
	return jarr
}

func (j *JSON) GetJSONArrayForInfluxDB(k string) []JSON {
	log.Println("inside getjsonarray")
	f := j.jmap[k]
	jarr := make([]JSON, 0)
	switch val := f.(type) {
	case []interface{}:
		for _, u := range val {
			switch val1 := u.(type) {
			case map[string]interface{}:
				log.Println("value of type: ", reflect.TypeOf(val))
				m := u.(map[string]interface{})
				jarr = append(jarr, NewFrom(m))

			case []interface{}:
				var jobj = New()
				for i, u1 := range val1 {
					key := ""
					if i == 0 {
						key = "time"
					} else if i == len(val1)-1 {
						key = "value"
					}
					switch u1.(type) {
					case string:
						jobj.Put(key, u1)
					case int:
						jobj.Put(key, u1.(int))
					case float64:
						jobj.Put(key, strconv.FormatFloat(float64(u1.(float64)), 'f', 3, 32))
					default:
						log.Println(u1)
						log.Println("value of type: ", reflect.TypeOf(u1))
					}

				}
				jarr = append(jarr, jobj)
			default:
				log.Println("value of type: ", reflect.TypeOf(u))
			}
		}
	}
	obj := New()
	obj.Put("data", jarr)
	log.Println("Jarray ", obj.ToString())
	return jarr
}

func (jobj *JSON) GetAsStringArray(k string) []string {
	v := jobj.jmap[k]
	str := make([]string, 0)
	switch val := v.(type) {
	case []float64: //float array
		for _, u := range val {
			str = append(str, fmt.Sprintf("%f", u))
		}
	case []int: //int array
		for _, u := range val {
			str = append(str, fmt.Sprintf("%d", u))
		}
	case []int32: //int array
		for _, u := range val {
			str = append(str, fmt.Sprintf("%d", u))
		}
	case []int64: //int array
		for _, u := range val {
			str = append(str, fmt.Sprintf("%d", u))
		}
	case []string: //String array
		for _, u := range val {
			str = append(str, fmt.Sprintf("%s", u))
		}
	case []bool: //String array
		for _, u := range val {
			str = append(str, fmt.Sprintf("%t`", u))
		}
	case []interface{}: //String array
		for _, u := range val {
			switch u.(type) {
			case float64:
				str = append(str, fmt.Sprintf("%f", u.(float64)))
			default:
				str = append(str, fmt.Sprintf("%s", u))
			}
		}
	case []JSON: //JSON array
		for _, u := range val {
			str = append(str, fmt.Sprintf("%s", u))
		}
	default:
		log.Println(k, "is of a type I don't know how to handle ", reflect.TypeOf(v))
	}
	return str
}

func (jobj *JSON) GetAsIntArray(k string) []int {
	v := jobj.jmap[k]
	str := make([]int, 0)
	switch val := v.(type) {
	case []float64: //float array
		for _, u := range val {
			str = append(str, int(u))
		}
	case []int: //int array
		for _, u := range val {
			str = append(str, u)
		}
	case []interface{}: //String array
		for _, u := range val {
			switch u.(type) {
			case float64:
				str = append(str, int(u.(float64)))
			default:
				str = append(str, u.(int))
			}
		}
	default:
		log.Println(k, "is of a type I don't know how to handle ", reflect.TypeOf(v))
	}
	return str
}

func (json *JSON) IsJSON(key string) bool {
	switch json.jmap[key].(type) {
	case map[string]interface{}, JSON:
		return true
	}
	return false
}

func (json *JSON) IsJSONArray(key string) bool {
	switch json.jmap[key].(type) {
	case []JSON:
		return true
	}
	return false
}

/////////////////////////////////////////////////////
////////////////////// ToString /////////////////////
/////////////////////////////////////////////////////

func ToJSONString(v interface{}) string {
	return string(ToJSONByte(v))
}

func ToJSONByte(v interface{}) []byte {
	buff, _ := json.Marshal(v)
	buff = bytes.Replace(buff, []byte("\\u0026"), []byte("&"), -1) //convert \u0026 into &
	return buff
}

func (jobj *JSON) ToString() string {
	m := jobj.jmap
	str := "{"
	for k, v := range m {

		switch val := v.(type) {

		case string:
			str += fmt.Sprintf("\"%s\":\"%s\",", k, val)
		case bool:
			str += fmt.Sprintf("\"%s\":%t,", k, val)
		case int:
			str += fmt.Sprintf("\"%s\":%d,", k, val)
		case int32:
			str += fmt.Sprintf("\"%s\":%d,", k, val)
		case int64:
			str += fmt.Sprintf("\"%s\":%d,", k, val)
		case float64:
			str += fmt.Sprintf("\"%s\":%f,", k, val)
		case JSON:
			str += fmt.Sprintf("\"%s\":%s,", k, val.ToString())
		case []float64: //float array
			str += fmt.Sprintf("\"%s\":[", k)
			for _, u := range val {
				str += fmt.Sprintf("%f,", u)
			}
			str = prune(str, ",") //prune extra comma if needed
			str += "],"
		case []int: //int array
			str += fmt.Sprintf("\"%s\":[", k)
			for _, u := range val {
				str += fmt.Sprintf("%d,", u)
			}
			str = prune(str, ",") //prune extra comma if needed
			str += "],"
		case []int64: //int array
			str += fmt.Sprintf("\"%s\":[", k)
			for _, u := range val {
				str += fmt.Sprintf("%d,", u)
			}
			str = prune(str, ",") //prune extra comma if needed
			str += "],"
		case []string: //String array
			str += fmt.Sprintf("\"%s\":[", k)
			for _, u := range val {
				str += fmt.Sprintf("\"%s\",", u)
			}
			str = prune(str, ",") //prune extra comma if needed
			str += "],"
		case []interface{}: //String array
			str += fmt.Sprintf("\"%s\":[", k)
			for _, u := range val {
				str += fmt.Sprintf("%s,", ToJSONString(u))
			}
			str = prune(str, ",") //prune extra comma if needed
			str += "],"
		case []JSON: //JSON array
			str += fmt.Sprintf("\"%s\":[", k)
			for _, u := range val {
				str += fmt.Sprintf("%s,", u.ToString())
			}
			str = prune(str, ",") //prune extra comma if needed
			str += "],"
		case map[string]interface{}: //JSON
			j := NewFrom(val)
			str += fmt.Sprintf("\"%s\":%s,", k, j.ToString())
		default:
			log.Println(k, "is of a type I don't know how to handle ", reflect.TypeOf(v))
		}
	}
	str = prune(str, ",") //prune extra comma if needed
	return str + "}"
}

func IsJSONValid(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func prune(str string, splitter string) string {
	if strings.LastIndex(str, splitter)+1 == len(str) {
		return str[0 : len(str)-1]
	}
	return str
}
