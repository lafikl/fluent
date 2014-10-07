package fluent

import (
	"bytes"
	"encoding/json"
	"github.com/cenkalti/backoff"
	"io"
	"net/http"
	"time"
)

type request struct {
	header    map[string]string
	method    string
	body      io.Reader
	json      interface{}
	jsonIsSet bool
	url       string
	retry     int
	timeout   time.Duration
	res       *http.Response
	err       error
	backoff   *backoff.ExponentialBackOff
}

func (f *request) newClient() *http.Client {
	return &http.Client{Timeout: f.timeout}
}

func (f *request) newRequest() (*http.Request, error) {
	var req *http.Request
	var err error
	if f.body != nil {
		req, err = http.NewRequest(f.method, f.url, f.body)
	} else if f.jsonIsSet {
		body, jsonErr := json.Marshal(f.json)
		if jsonErr != nil {
			return nil, jsonErr
		}
		req, err = http.NewRequest(f.method, f.url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(f.method, f.url, nil)
	}
	return req, err
}

func (f *request) Post(url string) *request {
	f.url = url
	f.method = "POST"
	return f
}

func (f *request) Put(url string) *request {
	f.url = url
	f.method = "PUT"
	return f
}

func (f *request) Get(url string) *request {
	f.url = url
	f.method = "GET"
	return f
}

func (f *request) Delete(url string) *request {
	f.url = url
	f.method = "DELETE"
	return f
}

func (f *request) Json(j interface{}) *request {
	f.json = j
	f.jsonIsSet = true
	f.SetHeader("Content-type", "application/json")
	return f
}

func (f *request) SetHeader(key, value string) *request {
	f.header[key] = value
	return f
}

func (f *request) Timeout(t time.Duration) *request {
	f.timeout = t
	return f
}

func (f *request) InitialInterval(t time.Duration) *request {
	f.backoff.InitialInterval = t
	return f
}

func (f *request) RandomizationFactor(rf float64) *request {
	f.backoff.RandomizationFactor = rf
	return f
}

func (f *request) Multiplier(m float64) *request {
	f.backoff.Multiplier = m
	return f
}

func (f *request) MaxInterval(mi time.Duration) *request {
	f.backoff.MaxInterval = mi
	return f
}

func (f *request) MaxElapsedTime(me time.Duration) *request {
	f.backoff.MaxElapsedTime = me
	return f
}

func (f *request) Clock(c backoff.Clock) *request {
	f.backoff.Clock = c
	return f
}

func (f *request) Retry(r int) *request {
	f.retry = r
	return f
}

func (f *request) operation(c *http.Client, req *http.Request) func() error {
	return func() error {
		res, err := c.Do(req)
		// if there's an error in the request
		// and there's no retries, then we just return whatever err we got
		if err != nil && f.retry <= 0 {
			f.err = err
			return nil
		} else if err != nil {
			// We still have retries, let's retry it!
			f.retry--
			return err
		} else if res.StatusCode >= 500 && res.StatusCode <= 599 {
			f.retry--
			return err
		}
		f.res = res
		return nil
	}
}

func (f *request) do(c *http.Client, req *http.Request) (*http.Response, error) {
	op := f.operation(c, req)

	err := backoff.Retry(op, f.backoff)
	if err != nil {
		return nil, err
	}
	// Check if has operation failed after the retries
	if f.err != nil {
		return nil, f.err
	}
	return f.res, err
}

func (f *request) Send() (*http.Response, error) {
	c := f.newClient()
	req, reqErr := f.newRequest()
	if reqErr != nil {
		return nil, reqErr
	}
	for k, v := range f.header {
		req.Header.Set(k, v)
	}
	res, err := f.do(c, req)
	return res, err
}

func New() *request {
	f := new(request)
	f.header = make(map[string]string)
	f.backoff = backoff.NewExponentialBackOff()
	f.err = nil
	return f
}
