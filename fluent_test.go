package fluent

import (
  "testing"
  "net/http"
  "net/http/httptest"
  "fmt"
  "io/ioutil"
  "time"
  "bytes"
  "strings"
)

func TestGet(t *testing.T) {
  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    fmt.Fprintln(w, string(body))
  }))
  defer ts.Close()

  req := New()
  res, err := req.Get(ts.URL).Send()
  if err != nil {
    t.Fatal(err)
  }

  if method := res.Request.Method; method != "GET" {
    t.Fatal("Method sent is not GET")
  }
}

func TestPost(t *testing.T) {
  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    fmt.Fprintln(w, string(body))
  }))
  defer ts.Close()

  req := New()
  res, err := req.Post(ts.URL).Send()
  if err != nil {
    t.Fatal(err)
  }

  if method := res.Request.Method; method != "POST" {
    t.Fatal("Method sent is not POST")
  }
}

func TestPut(t *testing.T) {
  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    fmt.Fprintln(w, string(body))
  }))
  defer ts.Close()

  req := New()
  res, err := req.Put(ts.URL).Send()
  if err != nil {
    t.Fatal(err)
  }

  if method := res.Request.Method; method != "PUT" {
    t.Fatal("Method sent is not PUT")
  }
}

func TestPatch(t *testing.T) {
  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    fmt.Fprintln(w, string(body))
  }))
  defer ts.Close()

  req := New()
  res, err := req.Patch(ts.URL).Send()
  if err != nil {
    t.Fatal(err)
  }

  if method := res.Request.Method; method != "PATCH" {
    t.Fatal("Method sent is not PATCH")
  }
}

func TestBody(t *testing.T) {
  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    fmt.Fprintf(w, string(body))
  }))
  defer ts.Close()

  msg := "Hello wld!"
  req := New()
  req.Post(ts.URL).
  Body(bytes.NewReader([]byte(msg))).
  Retry(2)
  res, err := req.Send()
  if err != nil {
    fmt.Println("errrrrr")
    t.Fatal(err)
  }
  body, err := ioutil.ReadAll(res.Body)
  res.Body.Close()
  if err != nil {
    t.Fatal(err)
  }
  b := string(body)
  b = strings.Trim(b, " \n")
  if b != msg {
    t.Fatalf("Body sent %s doesn't match %s", msg, b)
  }
}

func TestRetries(t *testing.T) {
  retry := 3
  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(500)
  }))
  defer ts.Close()

  req := New()
  req.Post(ts.URL).
  InitialInterval(time.Duration(time.Millisecond)).
  Json([]byte(`[1,2,3]`)).
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
