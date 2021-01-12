package misc

import (
	"encoding/json"
	"runtime"

	"github.com/gopaddle-io/sail/directory"
	gson "github.com/gopaddle-io/sail/util/json"
	"github.com/gopaddle-io/sail/util/log"
)

func BuildHTTPErrorJSON(message string, requestID string) string {
	errJSON, e := json.Marshal(Error{Message: message, RequestID: requestID})
	if e != nil {
		log.Errorln(e)
	}
	return string(errJSON)
}

// PanicHandler To handle server break from syntax errors
func PanicHandler(r interface{}, requestID string) Response {
	resp := Response{Code: 500}
	bytes := make([]byte, 2000)
	runtime.Stack(bytes, true)
	log.Errorf("Panic Message: %v and Error stack: %s", r, string(bytes))
	err := gson.New()
	err.Put("requestID", requestID)
	err.Put("reason", directory.ErrorFmt("All", "SOMETHING_WRONG"))
	resp.Response = err.ToString()
	switch x := r.(type) {
	case Error:
		err.Put("reason", x.Message)
		resp.Response = err.ToString()
	}

	return resp
}
