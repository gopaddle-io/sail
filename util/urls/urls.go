package urls

import (
	"bytes"
	"crypto/tls"
	gson "encoding/json"
	"fmt"
	"gopaddle/migrationservice/util/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Connection struct {
	URL     string
	Header  map[string]string
	Body    string
	TimtOut string
	Method  string
	Opaque  string
	Cookies []*http.Cookie
}

//Create new connection
func New() Connection {
	var con = &Connection{}
	con.Header = make(map[string]string)
	con.Header["Content-Type"] = "application/json"
	con.TimtOut = "0"
	con.Method = ""
	con.Cookies = nil
	return *con
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return gson.Unmarshal([]byte(s), &js) == nil

}

//To Execute based on HTTP Method
func (c *Connection) Execute() string {
	var code int
	var resp string
	var err error
	var res = json.New()
	if strings.ToLower(c.Method) == "get" {
		code, resp, err = c.Get()
	} else if strings.ToLower(c.Method) == "post" {
		code, resp, err = c.Post()
	} else if strings.ToLower(c.Method) == "put" {
		code, resp, err = c.Put()
	} else if strings.ToLower(c.Method) == "delete" {
		code, resp, err = c.Delete()
	} else {
		code, resp, err = c.Get()
	}
	res.Put("code", code)
	res.Put("response", json.ParseString(resp))
	if err == nil {
		res.Put("error", "")
	} else {
		res.Put("error", err.Error())
	}
	return res.ToString()
}

//GET Method
func (c *Connection) Get() (int, string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	// log.Println("ClientTimeout: ", client.Timeout)
	request, err_req := http.NewRequest("GET", c.URL, nil)
	if c.Opaque != "" {
		request.URL.Opaque = c.Opaque
		request.URL.Scheme = "https"
		request.URL.Host = "gitlab.com"
	}
	//Add cookies to header
	if c.Cookies != nil {
		for _, cookie := range c.Cookies {
			request.AddCookie(cookie)
		}
	}
	if err_req == nil {
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		} else {
			log.Println("Warn : Header is empty")
		} //Ensure header is present
		if response, err_response := client.Do(request); err_response == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, read_err := ioutil.ReadAll(response.Body); read_err == nil {
				return response.StatusCode, string(body), nil
			} else {
				log.Println("Error in reading http GET response : ", read_err)
				return response.StatusCode, "", read_err
			}
		} else {
			log.Println("Error in http GET connection : ", err_response)
			return 500, "", err_response
		}
	} else {
		log.Println("Error in new http GET method : ", err_req)
		return 500, "", err_req
	}
}

//GET Method
func (c *Connection) GetURLWithHeader() (int, string, string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	// log.Println("ClientTimeout: ", client.Timeout)
	request, err_req := http.NewRequest("GET", c.URL, nil)
	if err_req == nil {
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		} else {
			log.Println("Warn : Header is empty")
		} //Ensure header is present
		if response, err_response := client.Do(request); err_response == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, read_err := ioutil.ReadAll(response.Body); read_err == nil {
				header := response.Header.Get("X-RateLimit-Remaining")
				return response.StatusCode, string(body), header, nil
			} else {
				log.Println("Error in reading http GET response : ", read_err)
				return response.StatusCode, "", "", read_err
			}
		} else {
			log.Println("Error in http GET connection : ", err_response)
			return 500, "", "", err_response
		}
	} else {
		log.Println("Error in new http GET method : ", err_req)
		return 500, "", "", err_req
	}
}

//POST Method
func (c *Connection) Post() (int, string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	// log.Println("ClientTimeout: ", client.Timeout)
	var request *http.Request
	var err_req error
	if c.Body == "" {
		request, err_req = http.NewRequest("POST", c.URL, nil)
	} else {
		content_bytes := bytes.NewReader([]byte(c.Body))
		request, err_req = http.NewRequest("POST", c.URL, content_bytes)
	}
	//add cookie to header
	if c.Cookies != nil {
		for _, cookie := range c.Cookies {
			request.AddCookie(cookie)
		}
	}
	if err_req == nil {
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		} else {
			log.Println("Warn : Header is empty")
		} //Ensure header is present
		if response, err_response := client.Do(request); err_response == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, read_err := ioutil.ReadAll(response.Body); read_err == nil {
				return response.StatusCode, string(body), nil
			} else {
				log.Println("Error in reading http POST response : ", read_err)
				return response.StatusCode, "", read_err
			}
		} else {
			log.Println("Error in http POST connection : ", err_response)
			return 500, "", err_response
		}
	} else {
		log.Println("Error in new http POST method : ", err_req)
		return 500, "", err_req
	}
}

//POST Method which return Response with cookies
func (c *Connection) PostResponse() (int, *http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	// log.Println("ClientTimeout: ", client.Timeout)
	var request *http.Request
	var response *http.Response
	var err_req error
	if c.Body == "" {
		request, err_req = http.NewRequest("POST", c.URL, nil)
	} else {
		content_bytes := bytes.NewReader([]byte(c.Body))
		request, err_req = http.NewRequest("POST", c.URL, content_bytes)
	}
	if err_req == nil {
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		} else {
			log.Println("Warn : Header is empty")
		} //Ensure header is present
		if response, err_response := client.Do(request); err_response == nil {
			defer response.Body.Close() //Close connection when function ends
			return response.StatusCode, response, nil
		} else {
			log.Println("Error in http POST connection : ", err_response)
			return 500, response, err_response
		}
	} else {
		log.Println("Error in new http POST method : ", err_req)
		return 500, response, err_req
	}
}

//PUT METHOD
func (c *Connection) Put() (int, string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	// log.Println("ClientTimeout: ", client.Timeout)
	var request *http.Request
	var err_req error
	if c.Body == "" {
		request, err_req = http.NewRequest("PUT", c.URL, nil)
	} else {
		content_bytes := bytes.NewReader([]byte(c.Body))
		request, err_req = http.NewRequest("PUT", c.URL, content_bytes)
	}

	//add cookies if jira execution
	//add cookie to header
	if c.Cookies != nil {
		for _, cookie := range c.Cookies {
			//		log.Println("***", cookie, ")))))))))")
			request.AddCookie(cookie)
		}
	}
	if err_req == nil {
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		} else {
			log.Println("Warn : Header is empty")
		} //Ensure header is need to right now connection
		if response, err_response := client.Do(request); err_response == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, read_err := ioutil.ReadAll(response.Body); read_err == nil {
				return response.StatusCode, string(body), nil
			} else {
				log.Println("Error in reading http PUT response : ", read_err)
				return response.StatusCode, "", read_err
			}
		} else {
			log.Println("Error in http PUT connection : ", err_response)
			return 500, "", err_response
		}
	} else {
		log.Println("Error in new http  PUT method : ", err_req)
		return 500, "", err_req
	}
}

//DELETE Method
func (c *Connection) Delete() (int, string, error) {
	//The function will return status code
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	// log.Println("ClientTimeout: ", client.Timeout)
	var request *http.Request
	var err_req error
	if c.Body == "" {
		request, err_req = http.NewRequest("DELETE", c.URL, nil)
	} else {
		content_bytes := bytes.NewReader([]byte(c.Body))
		request, err_req = http.NewRequest("DELETE", c.URL, content_bytes)
	}

	//add cookie if it is jira request
	//add cookie to header
	if c.Cookies != nil {
		for _, cookie := range c.Cookies {
			//		log.Println("***", cookie, ")))))))))")
			request.AddCookie(cookie)
		}
	}
	if err_req == nil {
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		} else {
			log.Println("Warn : Header is empty")
		} //Ensure header is present
		if response, err_response := client.Do(request); err_response == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, read_err := ioutil.ReadAll(response.Body); read_err == nil {
				return response.StatusCode, string(body), nil
			} else {
				log.Println("Error in reading http DELETE response : ", read_err)
				return response.StatusCode, "", read_err
			}
		} else {
			log.Println("Error in http DELETE connection : ", err_response)
			return 500, "", err_response
		}
	} else {
		log.Println("Error in new http DELETE method : ", err_req)
		return 500, "", err_req
	}
}

func (c *Connection) HttpGetStream() (int, []byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	log.Println("ClientTimeout: ", client.Timeout)
	request, err_req := http.NewRequest("GET", c.URL, nil)
	if err_req == nil {
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		} else {
			log.Println("Warn : Header is empty")
		} //Ensure header is present
		if response, err_response := client.Do(request); err_response == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, read_err := ioutil.ReadAll(response.Body); read_err == nil {
				return response.StatusCode, body, nil
			} else {
				log.Println("Error in reading http GET response : ", read_err)
				return response.StatusCode, nil, read_err
			}
		} else {
			log.Println("Error in http GET connection : ", err_response)
			return 500, nil, err_response
		}
	} else {
		log.Println("Error in new http GET method : ", err_req)
		return 500, nil, err_req
	}
}

func FrameURL(host string, port int32, endpoint string, token string, isTLS bool) (map[string]string, string) {
	header := make(map[string]string)
	ssl := ""
	if isTLS {
		ssl = "s"
	}
	url_str := fmt.Sprintf("http%s://%s:%d/%s", ssl, host, port, endpoint)
	header["Authorization"] = "bearer " + token //H_6PvX5f3henTj0krvr1wx6FCCC5jXqRvR_BXCNUhfc
	return header, url_str
}

func EmitData(txURL, url, organizationID, state, reason, mac_name, vm_id, operation string) {
	//Form JSON data
	jobj := json.New()
	filteredJObj := json.New()
	jobj.Put("organizationID", organizationID)
	jobj.Put("operation", operation)
	jobj.Put("status", state)
	if state == "Failed" {
		jobj.Put("reason", reason)
	} else {
		if reason == "" {
			jobj.Put("reason", "Success")
		} else {
			jobj.Put("reason", "Success: "+reason)
		}
	}
	jobj.Put("name", mac_name)
	jobj.Put("vmid", vm_id)
	jobj.Put("created_time", fmt.Sprintf("%s", time.Now()))
	keyList := jobj.GetKeyList()
	for _, key := range keyList {
		if jobj.GetString(key) != "" {
			filteredJObj.Put(key, jobj.GetString(key))
		}
	}
	log.Println("Filter json:", filteredJObj)
	jobj1 := json.New()
	jobj1.Put("result", filteredJObj)
	result := jobj1.ToString()
	log.Println("Emit Data: ", result)
	// Call http method
	con := New()
	con.URL = url
	con.Body = result
	con.TimtOut = "2"
	if code, _, err := con.Post(); err != nil {
		log.Println("Error Emitting data", err)
	} else {
		if code == 200 {
			log.Println("Data emitted successfully")
		}
	}

	//Store in tx table
	var reqObj = json.New()
	reqObj.Put("organizationID", organizationID)
	reqObj.Put("jobID", vm_id)
	reqObj.Put("jobName", mac_name)
	reqObj.Put("operation", operation)
	reqObj.Put("reason", reason)
	reqObj.Put("status", state)
	con1 := New()
	con1.URL = txURL
	con1.Body = reqObj.ToString()
	con1.TimtOut = "2"
	if code1, _, err1 := con1.Post(); err1 != nil {
		log.Println("Error saving transaction: ", err1)
	} else {
		if code1 == 200 {
			log.Println("Transaction saved successfully")
		}
	}
}

func EmitDataForORganization(url string, state string, reason string, mac_name string, vm_id string, operation string, message string) {
	//Declare header
	var header = make(map[string]string)
	header["Content-Type"] = "application/json" //Declared globally
	//Form JSON data
	jobj := json.New()
	filteredJObj := json.New()
	jobj.Put("operation", operation)
	jobj.Put("status", state)
	jobj.Put("message", message)
	if state == "Failed" {
		jobj.Put("reason", reason)
	} else {
		jobj.Put("reason", "")
	}
	jobj.Put("name", mac_name)
	jobj.Put("vmid", vm_id)
	jobj.Put("created_time", fmt.Sprintf("%s", time.Now()))
	keyList := jobj.GetKeyList()
	for _, key := range keyList {
		if jobj.GetString(key) != "" {
			filteredJObj.Put(key, jobj.GetString(key))
		}
	}
	log.Println("Filter json:", filteredJObj)
	jobj1 := json.New()
	jobj1.Put("result", filteredJObj)
	result := jobj1.ToString()
	log.Println("Emit Data: ", result)
	//Call http method
	// code, _, err := HttpPOST(url, header, result)
	// if err != nil {
	// 	log.Println("Error Emitting data", err)
	// } else {
	// 	if code == 200 {
	// 		log.Println("Data emitted successfully")
	// 	}
	// }
}
