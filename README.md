go-fluent
=========

Fluent HTTP client for Golang. With timeout, retries and exponential back-off support.

Usage:

```go
package main

import (
  "github.com/lafikl/go-fluent"
  "fmt"
)

func main() {
  req := fluent.New()
  req.Get("http://example.com")
  req.Retry(3)
  // They're chaniable too 
  // req.Get("http://example.com").Retry(3)

  res, err := req.Send()

  if err != nil {
    fmt.Println(err)
  }

  fmt.Println("donne ", res)
}

```

http://godoc.org/github.com/lafikl/go-fluent


**NOTE:** I'm still new to Go so if you find anything that isn't Go idiomatic, please open an issue :+1: