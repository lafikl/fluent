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

func TestDelete(t *testing.T) {
	ts := httptest.NewServer(copyHandlerFunc)
	defer ts.Close()

	res, err := New().Delete(ts.URL).Send()
	if err != nil {
		t.Fatal(err)
	}

	if res.Request.Method != "DELETE" {
		t.Fatal("Method sent is not DELETE")
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

func TestTimeout(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}),
	)
	defer ts.Close()
	req := New()
	req.Get(ts.URL).Timeout(time.Duration(time.Second)).Send()
	if req.timeout == 0 {
		t.Fatal("timeout should be more than zero")
	}

	c := req.newClient()
	if c.Timeout == 0 {
		t.Fatal("Client timeout should be more than zero")
	}
}

func TestRandomizationFactor(t *testing.T) {
	req := New()
	req.RandomizationFactor(0.6)
	// 0.5 is the default that's why i'm testing against it
	if req.backoff.RandomizationFactor != 0.6 {
		t.Fatal("RandomizationFactor should be 0.6")
	}
}

func TestMultiplier(t *testing.T) {
	req := New()
	req.Multiplier(2.0)
	if req.backoff.Multiplier != 2.0 {
		t.Fatal("Multiplier should be 2.0")
	}
}

func TestMaxInterval(t *testing.T) {
	interval := time.Duration(20 * time.Second)
	req := New()
	req.MaxInterval(interval)
	if req.backoff.MaxInterval != interval {
		t.Fatalf("MaxInterval should be %s", interval)
	}
}

func TestMaxElapsedTime(t *testing.T) {
	elapsed := time.Duration(20 * time.Second)
	req := New()
	req.MaxElapsedTime(elapsed)
	if req.backoff.MaxElapsedTime != elapsed {
		t.Fatalf("MaxElapsedTime should be %s", elapsed)
	}
}
