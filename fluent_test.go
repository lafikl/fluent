package fluent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func readAllString(r io.Reader) (string, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var copyHandlerFunc = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		io.Copy(w, r.Body)
	},
)

func TestGet(t *testing.T) {
	ts := httptest.NewServer(copyHandlerFunc)
	defer ts.Close()

	res, err := New().Get(ts.URL).Send()
	if err != nil {
		t.Fatal(err)
	}

	if method := res.Request.Method; method != "GET" {
		t.Fatal("Method sent is not GET")
	}
}

func TestPost(t *testing.T) {
	ts := httptest.NewServer(copyHandlerFunc)
	defer ts.Close()

	res, err := New().Post(ts.URL).Send()
	if err != nil {
		t.Fatal(err)
	}

	if res.Request.Method != "POST" {
		t.Fatal("Method sent is not POST")
	}
}

func TestPut(t *testing.T) {
	ts := httptest.NewServer(copyHandlerFunc)
	defer ts.Close()

	res, err := New().Put(ts.URL).Send()
	if err != nil {
		t.Fatal(err)
	}

	if res.Request.Method != "PUT" {
		t.Fatal("Method sent is not PUT")
	}
}

func TestPatch(t *testing.T) {
	ts := httptest.NewServer(copyHandlerFunc)
	defer ts.Close()

	res, err := New().Patch(ts.URL).Send()
	if err != nil {
		t.Fatal(err)
	}

	if res.Request.Method != "PATCH" {
		t.Fatal("Method sent is not PATCH")
	}
}

func TestBody(t *testing.T) {
	ts := httptest.NewServer(copyHandlerFunc)
	defer ts.Close()

	msg := "Hello world!"
	res, err := New().
		Post(ts.URL).
		Body(strings.NewReader(msg)).
		Send()
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := readAllString(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if body != msg {
		t.Fatalf("Body sent %s doesn't match %s", msg, body)
	}
}

func TestJson(t *testing.T) {
	ts := httptest.NewServer(copyHandlerFunc)
	defer ts.Close()

	arr := []int{1, 2, 3}
	res, err := New().Post(ts.URL).Json(arr).Send()
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := readAllString(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if body != "[1,2,3]" {
		t.Fatalf("JSON sent doesn't match %s", body)
	}
}

func TestRetries(t *testing.T) {
	retry := 3
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}),
	)
	defer ts.Close()

	req := New()
	req.Post(ts.URL).
		InitialInterval(time.Millisecond).
		Json([]int{1, 3, 4}).
		Retry(retry)
	if req.retry != retry {
		t.Fatalf("Retries didn't apply!")
	}
	_, err := req.Send()

	if err != nil {
		fmt.Println("err", err)
	}

	if req.retry != 0 {
		t.Fatalf("Fluent exited without finishing retries")
	}
}
