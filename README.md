# compress
[![Build Status](https://travis-ci.org/go-http-utils/compress.svg?branch=master)](https://travis-ci.org/go-http-utils/compress)
[![Coverage Status](https://coveralls.io/repos/github/go-http-utils/compress/badge.svg?branch=master)](https://coveralls.io/github/go-http-utils/compress?branch=master)

Compress middleware for Go.

## Installation

```
go get -u github.com/go-http-utils/compress
```

## Documentation

API documentation can be found here: https://godoc.org/github.com/go-http-utils/compress

## Usage

```go
import (
  "github.com/go-http-utils/compress"
)
```

```go
mux := http.NewServeMux()
mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
  res.Write([]byte("Hello World"))
})

http.ListenAndServe(":8080", compress.Handler(mux))
```
