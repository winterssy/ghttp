package ghttp

import (
	"net/http"
	neturl "net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpRequestHeaders(t *testing.T) {
	dummyRequest := &Request{Request: &http.Request{URL: &neturl.URL{
		Scheme: "https",
		Host:   "httpbin.org",
		Path:   "/get",
	}}}
	dummyRequest.TransferEncoding = []string{"chunked"}
	dummyRequest.Close = true

	var sb strings.Builder
	dumpRequestHeaders(dummyRequest, &sb)
	want := "" +
		"> Host: httpbin.org\r\n" +
		"> Transfer-Encoding: chunked\r\n" +
		"> Connection: close\r\n" +
		">\r\n"
	assert.Equal(t, want, sb.String())
}
