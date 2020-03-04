package ghttp

import (
	"fmt"
	"io"
)

type (
	debugger struct {
		out  io.Writer
		body bool
	}
)

// Enter implements BeforeRequestCallback interface.
func (d *debugger) Enter(req *Request) error {
	err := dumpRequest(req, d.out, d.body)
	if err != nil {
		fmt.Fprintf(d.out, "* ghttp [ERROR] %s\r\n", err.Error())
	}
	return err
}

// Exit implements AfterResponseCallback interface.
func (d *debugger) Exit(resp *Response, err error) {
	if err == nil {
		err = dumpResponse(resp, d.out, d.body)
	}
	if err != nil {
		fmt.Fprintf(d.out, "* ghttp [ERROR] %s\r\n", err.Error())
	}
}
