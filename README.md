# ghttp

**[ghttp](https://pkg.go.dev/github.com/winterssy/ghttp)** is a simple, user-friendly and concurrent safe HTTP request library for Go.

![Build](https://img.shields.io/github/workflow/status/winterssy/ghttp/Test/master?logo=appveyor) [![codecov](https://codecov.io/gh/winterssy/ghttp/branch/master/graph/badge.svg)](https://codecov.io/gh/winterssy/ghttp) [![Go Report Card](https://goreportcard.com/badge/github.com/winterssy/ghttp)](https://goreportcard.com/report/github.com/winterssy/ghttp) [![GoDoc](https://img.shields.io/badge/godoc-reference-5875b0)](https://pkg.go.dev/github.com/winterssy/ghttp) [![License](https://img.shields.io/github/license/winterssy/ghttp.svg)](LICENSE)

## Features

`ghttp` wraps `net/http` and provides convenient APIs and advanced features to simplify your jobs.

- Requests-style APIs.
- GET, POST, PUT, PATCH, DELETE, etc.
- Easy set query params, headers and cookies.
- Easy send form, JSON or multipart payload.
- Automatic cookies management.
- Backoff retry mechanism.
- Before request and after response callbacks.
- Rate limiting for outbound requests.
- Easy decode the response body to bytes, string or unmarshal the JSON-encoded data.
- Friendly debugging.
- Concurrent safe.

## Install

```sh
go get -u github.com/winterssy/ghttp
```

## Usage

```go
import "github.com/winterssy/ghttp"
```

## Quick Start

The usages of `ghttp` are very similar to `net/http` .

- `ghttp.Client`

```go
client := ghttp.New()
// Now you can manipulate client like net/http
client.CheckRedirect = ghttp.NoRedirect
client.Timeout = 300 * time.Second
```

- `ghttp.Request`

```go
req, err := ghttp.NewRequest("GET", "https://httpbin.org/get")
if err != nil {
    log.Fatal(err)
}
// Now you can manipulate req like net/http
req.Close = true
```

- `ghttp.Response`

```go
resp, err := ghttp.Get("https://www.google.com")
if err != nil {
    log.Fatal(err)
}
// Now you can access resp like net/http
fmt.Println(resp.StatusCode)
```

Documentation is available at  **[go.dev](https://pkg.go.dev/github.com/winterssy/ghttp)** .

## License

**[MIT](LICENSE)**
