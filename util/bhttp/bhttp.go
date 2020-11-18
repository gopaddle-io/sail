package bhttp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"gopaddle/migrationservice/util"
	"gopaddle/migrationservice/util/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryHTTP struct
type retryHTTP struct {
	attempt int
	args    httpConnection
	backOff int
	delay   int
	fn      func(httpConnection) HTTPResponse
	log     *logrus.Entry
}

// Retry It will retry all failure calls for the given certain of times
func Retry(f func(httpConnection) HTTPResponse, args httpConnection, log *logrus.Entry) *retryHTTP {

	return &retryHTTP{
		fn:      f,
		args:    args,
		delay:   30,
		attempt: 3,
		backOff: 1,
		log:     log,
	}
}

// WithAttempt It allows to change your default retry attempt count
func (r *retryHTTP) WithAttempt(attempt int) *retryHTTP {
	r.attempt = attempt
	return r
}

// WithBackOff It allows to change your default retry backoff count
func (r *retryHTTP) WithBackOff(backOff int) *retryHTTP {
	r.backOff = backOff
	return r
}

// WithDelay It allows to change your default delay count
func (r *retryHTTP) WithDelay(delay int) *retryHTTP {
	r.delay = delay
	return r
}

// Do It does executes the function
func (r *retryHTTP) Do() HTTPResponse {
	var resp HTTPResponse
	for i := 1; i <= r.attempt; i++ {
		if resp = r.fn(r.args); !resp.retryNeed {
			break
		}
		delay := (r.backOff * (i - 1)) + r.delay
		r.log.Warnf("Retrying %d attempt within %d seconds", i, delay)
		time.Sleep(time.Second * time.Duration(delay))
	}
	return resp
}

type httpConnection struct {
	URL          string
	Header       map[string]string
	Body         string
	TimtOut      string
	Opaque       string
	Cookies      []*http.Cookie
	ExpectedCode []int
	log          *logrus.Entry
}

// NewHTTPConnection Default configuration for http connection
func NewHTTPConnection(url string) *httpConnection {
	var con = &httpConnection{}
	con.URL = url
	con.Header = make(map[string]string)
	con.Header["Content-Type"] = "application/json"
	con.TimtOut = "10"
	con.Cookies = nil
	con.ExpectedCode = []int{200, 202}
	return con
}

// HTTPInterface It handles all kind of rest calls
type HTTPInterface interface {
	// To handle http get calls
	Get(httpConnection) HTTPResponse

	// To handle http post calls
	Post(httpConnection) HTTPResponse

	// To handle http put calls
	Put(httpConnection) HTTPResponse

	// To handle http delete calls
	Delete(httpConnection) HTTPResponse
}

type httpCaller struct {
	log *logrus.Entry
}

// NewHTTPCaller HTTPCaller's constructor
func NewHTTPCaller(log *logrus.Entry) *httpCaller {
	return &httpCaller{log: log}
}

// Get To handle http GET method calls
func (caller *httpCaller) Get(h httpConnection) HTTPResponse {
	h.log = caller.log
	return h.get()
}

// Put To handle http PUT method calls
func (caller *httpCaller) Put(h httpConnection) HTTPResponse {
	h.log = caller.log
	return h.put()
}

// Post To handle http POST method calls
func (caller *httpCaller) Post(h httpConnection) HTTPResponse {
	h.log = caller.log
	return h.post()
}

// Delete To handle http DELETE method calls
func (caller *httpCaller) Delete(h httpConnection) HTTPResponse {
	h.log = caller.log
	return h.delete()
}

// HTTPResponse Response from http calls
type HTTPResponse struct {
	retryNeed      bool
	err            error
	Response       string
	ActualResponse ActualResponse
}

// ActualResponse To handle if user given urls
type ActualResponse struct {
	Error      error
	StatusCode int
	Body       string
}

// IsSucceeded Check that made the call is scceeded or not
func (httpResponse *HTTPResponse) IsSucceeded() bool {
	if httpResponse.err == nil {
		return true
	}
	return false
}

// GetError It will Parse error from http response
// It returns parsed error message if response contains either reason or message
// otherwise it returns entire message
func (httpResponse *HTTPResponse) GetError() string {
	if httpResponse.err != nil {
		if json.IsJSONValid(httpResponse.Response) {
			jsonResp := json.ParseString(httpResponse.Response)
			if jsonResp.HasKey("reason") {
				return jsonResp.GetString("reason")
			} else if jsonResp.HasKey("message") {
				return jsonResp.GetString("message")
			} else {
				return httpResponse.Response
			}
		}
		return httpResponse.err.Error()
	}
	return ""
}

func (httpResponse *HTTPResponse) JSON() *json.JSON {
	j := json.ParseString(httpResponse.Response)
	return &j
}

func (httpResponse *HTTPResponse) ActualRespToStr() string {
	return fmt.Sprintf("Code: %d, Response: %s, Error: %v",
		httpResponse.ActualResponse.StatusCode,
		httpResponse.ActualResponse.Body,
		httpResponse.ActualResponse.Error,
	)
}

// CodeConverter Convert actual code into appropriate user request
func (httpResponse *HTTPResponse) CodeConverter() int {
	switch httpResponse.ActualResponse.StatusCode {
	case 500, -1:
		return 500
	case 400, 404:
		return 400
	}
	return httpResponse.ActualResponse.StatusCode
}

func (c *httpConnection) get() HTTPResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	request, err := http.NewRequest("GET", c.URL, nil)
	if err == nil {
		if c.Opaque != "" {
			request.URL.Opaque = c.Opaque
			request.URL.Scheme = "https"
		}
		//Add cookies to header
		if c.Cookies != nil {
			for _, cookie := range c.Cookies {
				request.AddCookie(cookie)
			}
		}
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		}

		if response, errResponse := client.Do(request); errResponse == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, readErr := ioutil.ReadAll(response.Body); readErr == nil {
				if util.IntContains(response.StatusCode, c.ExpectedCode) {
					return HTTPResponse{
						Response: string(body),
					}
				}
				c.log.Errorf("Invalid http response code: %d, %s", response.StatusCode, string(body))
				return HTTPResponse{
					err:      errors.New("Something went wrong! please try again later"),
					Response: string(body),
					ActualResponse: ActualResponse{
						Body:       string(body),
						StatusCode: response.StatusCode,
					},
				}
			} else {
				c.log.Errorln(readErr)
				return HTTPResponse{
					retryNeed: true,
					err:       errors.New("Something went wrong! please try again later"),
					ActualResponse: ActualResponse{
						Error:      readErr,
						StatusCode: -1,
					},
				}
			}
		} else {
			c.log.Errorln(errResponse)
			return HTTPResponse{
				retryNeed: true,
				err:       errors.New("Something went wrong! please try again later"),
				ActualResponse: ActualResponse{
					Error:      errResponse,
					StatusCode: -1,
				},
			}
		}
	} else {
		c.log.Errorln(err)
		return HTTPResponse{
			retryNeed: true,
			err:       errors.New("Something went wrong! please try again later"),
			ActualResponse: ActualResponse{
				Error:      err,
				StatusCode: -1,
			},
		}
	}
}

// post To handle http POST method calls
func (c *httpConnection) post() HTTPResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	var request *http.Request
	var err error
	if c.Body == "" {
		request, err = http.NewRequest("POST", c.URL, nil)
	} else {
		request, err = http.NewRequest("POST", c.URL, bytes.NewReader([]byte(c.Body)))
	}

	if err == nil {
		//Add cookies to header
		if c.Cookies != nil {
			for _, cookie := range c.Cookies {
				request.AddCookie(cookie)
			}
		}
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		}

		if response, errResponse := client.Do(request); errResponse == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, readErr := ioutil.ReadAll(response.Body); readErr == nil {
				if util.IntContains(response.StatusCode, c.ExpectedCode) {
					return HTTPResponse{
						Response: string(body),
					}
				}
				c.log.Errorf("Invalid http response code: %d, %s", response.StatusCode, string(body))
				return HTTPResponse{
					err:      errors.New("Something went wrong! please try again later"),
					Response: string(body),
					ActualResponse: ActualResponse{
						Body:       string(body),
						StatusCode: response.StatusCode,
					},
				}
			} else {
				c.log.Errorln(readErr)
				return HTTPResponse{
					retryNeed: true,
					err:       errors.New("Something went wrong! please try again later"),
					ActualResponse: ActualResponse{
						Error:      readErr,
						StatusCode: -1,
					},
				}
			}
		} else {
			c.log.Errorln(errResponse)
			return HTTPResponse{
				retryNeed: true,
				err:       errors.New("Something went wrong! please try again later"),
				ActualResponse: ActualResponse{
					Error:      errResponse,
					StatusCode: -1,
				},
			}
		}
	} else {
		c.log.Errorln(err)
		return HTTPResponse{
			retryNeed: true,
			err:       errors.New("Something went wrong! please try again later"),
			ActualResponse: ActualResponse{
				Error:      err,
				StatusCode: -1,
			},
		}
	}
}

// put To handle http PUT method calls
func (c *httpConnection) put() HTTPResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	var request *http.Request
	var err error
	if c.Body == "" {
		request, err = http.NewRequest("PUT", c.URL, nil)
	} else {
		request, err = http.NewRequest("PUT", c.URL, bytes.NewReader([]byte(c.Body)))
	}

	if err == nil {
		//Add cookies to header
		if c.Cookies != nil {
			for _, cookie := range c.Cookies {
				request.AddCookie(cookie)
			}
		}
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		}

		if response, errResponse := client.Do(request); errResponse == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, readErr := ioutil.ReadAll(response.Body); readErr == nil {
				if util.IntContains(response.StatusCode, c.ExpectedCode) {
					return HTTPResponse{
						Response: string(body),
					}
				}
				c.log.Errorf("Invalid http response code: %d, %s", response.StatusCode, string(body))
				return HTTPResponse{
					err:      errors.New("Something went wrong! please try again later"),
					Response: string(body),
					ActualResponse: ActualResponse{
						Body:       string(body),
						StatusCode: response.StatusCode,
					},
				}
			} else {
				c.log.Errorln(readErr)
				return HTTPResponse{
					retryNeed: true,
					err:       errors.New("Something went wrong! please try again later"),
					ActualResponse: ActualResponse{
						Error:      readErr,
						StatusCode: -1,
					},
				}
			}
		} else {
			c.log.Errorln(errResponse)
			return HTTPResponse{
				retryNeed: true,
				err:       errors.New("Something went wrong! please try again later"),
				ActualResponse: ActualResponse{
					Error:      errResponse,
					StatusCode: -1,
				},
			}
		}
	} else {
		c.log.Errorln(err)
		return HTTPResponse{
			retryNeed: true,
			err:       errors.New("Something went wrong! please try again later"),
			ActualResponse: ActualResponse{
				Error:      err,
				StatusCode: -1,
			},
		}
	}
}

// delete To handle http DELETE method calls
func (c *httpConnection) delete() HTTPResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if c.TimtOut != "0" {
		timer, _ := time.ParseDuration(c.TimtOut + "s")
		client = &http.Client{Transport: tr, Timeout: timer}
	}
	var request *http.Request
	var err error
	if c.Body == "" {
		request, err = http.NewRequest("DELETE", c.URL, nil)
	} else {
		request, err = http.NewRequest("DELETE", c.URL, bytes.NewReader([]byte(c.Body)))
	}

	if err == nil {
		//Add cookies to header
		if c.Cookies != nil {
			for _, cookie := range c.Cookies {
				request.AddCookie(cookie)
			}
		}
		request.Header.Set("Connection", "close") //Ensure no persistant connection
		if c.Header != nil {
			for key, value := range c.Header {
				request.Header.Set(key, value)
			}
		}

		if response, errResponse := client.Do(request); errResponse == nil {
			defer response.Body.Close() //Close connection when function ends
			if body, readErr := ioutil.ReadAll(response.Body); readErr == nil {
				if util.IntContains(response.StatusCode, c.ExpectedCode) {
					return HTTPResponse{
						Response: string(body),
					}
				}
				c.log.Errorf("Invalid http response code: %d, %s", response.StatusCode, string(body))
				return HTTPResponse{
					err:      errors.New("Something went wrong! please try again later"),
					Response: string(body),
					ActualResponse: ActualResponse{
						Body:       string(body),
						StatusCode: response.StatusCode,
					},
				}
			} else {
				c.log.Errorln(readErr)
				return HTTPResponse{
					retryNeed: true,
					err:       errors.New("Something went wrong! please try again later"),
					ActualResponse: ActualResponse{
						Error:      readErr,
						StatusCode: -1,
					},
				}
			}
		} else {
			c.log.Errorln(errResponse)
			return HTTPResponse{
				retryNeed: true,
				err:       errors.New("Something went wrong! please try again later"),
				ActualResponse: ActualResponse{
					Error:      errResponse,
					StatusCode: -1,
				},
			}
		}
	} else {
		c.log.Errorln(err)
		return HTTPResponse{
			retryNeed: true,
			err:       errors.New("Something went wrong! please try again later"),
			ActualResponse: ActualResponse{
				Error:      err,
				StatusCode: -1,
			},
		}
	}
}
