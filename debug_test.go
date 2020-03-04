package ghttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugger_Enter(t *testing.T) {
	dummyRequest := &Request{
		Request: &http.Request{
			Method: MethodPost,
			URL: &neturl.URL{
				Scheme: "https",
				Host:   "httpbin.org",
				Path:   "/post",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
			Body:       &dummyBody{errFlag: errRead},
			Host:       "httpbin.org",
		},
	}
	debugger := &debugger{out: ioutil.Discard, body: true}
	err := debugger.Enter(dummyRequest)
	assert.Equal(t, errAccessDummyBody, err)
}

func TestDebugger_Exit(t *testing.T) {
	var sb strings.Builder
	debugger := &debugger{out: &sb, body: true}
	var dummyResponse *Response
	debugger.Exit(dummyResponse, errAccessDummyBody)
	assert.Equal(t, fmt.Sprintf("* ghttp [ERROR] %s\r\n", errAccessDummyBody), sb.String())
}
