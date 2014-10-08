package fluent

import (
	"bytes"
	"encoding/json"
	"github.com/lafikl/backoff"
	"net/http"
	"time"
	"errors"
	"io"
)

type Request struct {
	header    map[string]string
	method    string
	json      interface{}
	jsonIsSet bool
	url       string
	retry     int
	timeout   time.Duration
	body      io.Reader
	res       *http.Response
	err       error
	backoff   *backoff.ExponentialBackOff
	req 			*http.Request
}

func (f *Request) newClient() *http.Client {
	return &http.Client{Timeout: f.timeout}
}

func (f *Request) newRequest() (*http.Request, error) {
	var req *http.Request
	var err error
	if f.jsonIsSet {
		body, jsonErr := json.Marshal(f.json)
		if jsonErr != nil {
			return nil, jsonErr
		}
		req, err = http.NewRequest(f.method, f.url, bytes.NewReader(body))
	} else if f.body != nil {
		req, err = http.NewRequest(f.method, f.url, f.body)
	} else {
		req, err = http.NewRequest(f.method, f.url, nil)
	}
	return req, err
}

// Set the request URL
// You probably want to use the methods [Post, Get, Patch, Delete, Put]
func (f *Request) Url(url string) *Request {
	f.url = url
	return f
}

// Set the request Method
// You probably want to use the methods [Post, Get, Patch, Delete, Put]
func (f *Request) Method(method string) *Request {
	f.method = method
	return f
}

// This is a shorthand method that calls f.Method with `POST`
// and calls f.Url with the url you give to her 
func (f *Request) Post(url string) *Request {
	f.Url(url).Method("POST")
	return f
}

// Same as f.Post but the method is `PUT`
func (f *Request) Put(url string) *Request {
	f.Url(url).Method("PUT")
	return f
}

// Same as f.Post but the method is `PATCH`
func (f *Request) Patch(url string) *Request {
	f.Url(url).Method("PATCH")
	return f
}

// Same as f.Post but the method is `GET`
func (f *Request) Get(url string) *Request {
	f.Url(url).Method("GET")
	return f
}

// Same as f.Post but the method is `DELETE`
func (f *Request) Delete(url string) *Request {
	f.Url(url).Method("DELETE")
	return f
}

// A handy method for sending json without needing to Marshal it yourself
// This method will override whatever you pass to f.Body
// And it sets the content-type to "application/json" 
func (f *Request) Json(j interface{}) *Request {
	f.json = j
	f.jsonIsSet = true
	f.SetHeader("Content-type", "application/json")
	return f
}

// Whatever you pass to it will be passed to http.NewRequest
func (f *Request) Body(b io.Reader) *Request {
	f.body = b
	return f
}

// sets the header entries associated with key to the element value.
// 
// It replaces any existing values associated with key.
func (f *Request) SetHeader(key, value string) *Request {
	f.header[key] = value
	return f
}

// Timeout specifies a time limit for requests made by this
// Client. The timeout includes connection time, any
// redirects, and reading the response body. The timer remains
// running after Get, Head, Post, or Do return and will
// interrupt reading of the Response.Body.
//
// A Timeout of zero means no timeout.
func (f *Request) Timeout(t time.Duration) *Request {
	f.timeout = t
	return f
}

// The initial interval for the request backoff operation
// the default is `500 * time.Millisecond`
func (f *Request) InitialInterval(t time.Duration) *Request {
	f.backoff.InitialInterval = t
	return f
}

func (f *Request) RandomizationFactor(rf float64) *Request {
	f.backoff.RandomizationFactor = rf
	return f
}

func (f *Request) Multiplier(m float64) *Request {
	f.backoff.Multiplier = m
	return f
}

func (f *Request) MaxInterval(mi time.Duration) *Request {
	f.backoff.MaxInterval = mi
	return f
}

func (f *Request) MaxElapsedTime(me time.Duration) *Request {
	f.backoff.MaxElapsedTime = me
	return f
}

func (f *Request) Clock(c backoff.Clock) *Request {
	f.backoff.Clock = c
	return f
}

func (f *Request) Retry(r int) *Request {
	f.retry = r
	return f
}

func doReq(f *Request, c *http.Client) error {
	var reqErr error
	f.req, reqErr = f.newRequest()
	if reqErr != nil {
		return reqErr
	}
	for k, v := range f.header {
		f.req.Header.Set(k, v)
	}
	res, err := c.Do(f.req)
	// if there's an error in the request
	// and there's no retries, then we just return whatever err we got
	if err != nil {
		f.err = err
		return nil
	}
	if res != nil && res.StatusCode >= 500 && res.StatusCode <= 599 && f.retry > 0 {
		f.retry--
		return errors.New("Server Error")
	}
	if res != nil {
		f.res = res	
	}
	return nil
}

func (f *Request) operation(c *http.Client) func() error {
	return func() error {
		return doReq(f, c)
	}
}

func (f *Request) do(c *http.Client) (*http.Response, error) {
	err := doReq(f, c)
	if err != nil {
			op := f.operation(c)
			err = backoff.Retry(op, f.backoff)
			if err != nil {
				return nil, err
			}
	}
	// Check if has operation failed after the retries
	if f.err != nil {
		return nil, f.err
	}
	return f.res, err
}

func (f *Request) Send() (*http.Response, error) {
	c := f.newClient()
	res, err := f.do(c)
	return res, err
}

func New() *Request {
	f := &Request{}
	f.header = map[string]string{}
	f.backoff = backoff.NewExponentialBackOff()
	f.err = nil
	return f
}
