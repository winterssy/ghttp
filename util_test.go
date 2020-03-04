package ghttp

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	errRead = 1 << iota
	errClose
)

var (
	errAccessDummyBody = errors.New("dummy body is inaccessible")
)

type (
	dummyBody struct {
		s       string
		i       int64
		errFlag int
	}
)

func (db *dummyBody) Read(b []byte) (n int, err error) {
	if db.errFlag&errRead != 0 {
		return 0, errAccessDummyBody
	}

	if db.i >= int64(len(db.s)) {
		return 0, io.EOF
	}

	n = copy(b, db.s[db.i:])
	db.i += int64(n)
	return
}

func (db *dummyBody) Close() error {
	if db.errFlag&errClose != 0 {
		return errAccessDummyBody
	}

	return nil
}

func TestToString(t *testing.T) {
	var (
		stringVal             = "hi"
		bytesVal              = []byte{'h', 'e', 'l', 'l', 'o'}
		bytesValStr           = "hello"
		boolValStr            = "true"
		intVal                = -314
		intValStr             = "-314"
		int8Val       int8    = -128
		int8ValStr            = "-128"
		int16Val      int16   = -32768
		int16ValStr           = "-32768"
		int32Val      int32   = -314159
		int32ValStr           = "-314159"
		int64Val      int64   = -31415926535
		int64ValStr           = "-31415926535"
		uintVal       uint    = 314
		uintValStr            = "314"
		uint8Val      uint8   = 127
		uint8ValStr           = "127"
		uint16Val     uint16  = 32767
		uint16ValStr          = "32767"
		uint32Val     uint32  = 314159
		uint32ValStr          = "314159"
		uint64Val     uint64  = 31415926535
		uint64ValStr          = "31415926535"
		float32Val    float32 = 3.14159
		float32ValStr         = "3.14159"
		float64Val            = 3.1415926535
		float64ValStr         = "3.1415926535"
		errVal                = errAccessDummyBody
		timeVal               = time.Now()
	)
	tests := []struct {
		input interface{}
		want  string
	}{
		{stringVal, stringVal},
		{bytesVal, bytesValStr},
		{true, boolValStr},
		{intVal, intValStr},
		{int8Val, int8ValStr},
		{int16Val, int16ValStr},
		{int32Val, int32ValStr},
		{int64Val, int64ValStr},
		{uintVal, uintValStr},
		{uint8Val, uint8ValStr},
		{uint16Val, uint16ValStr},
		{uint32Val, uint32ValStr},
		{uint64Val, uint64ValStr},
		{float32Val, float32ValStr},
		{float64Val, float64ValStr},
		{errVal, errVal.Error()},
		{timeVal, timeVal.String()},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, toString(test.input))
	}

	complexVal := 1 + 2i
	assert.Panics(t, func() {
		toString(complexVal)
	})
}

func TestToStrings(t *testing.T) {
	vs := []string{"1", "2", "3"}

	var v interface{} = []int{1, 2, 3}
	assert.Equal(t, vs, toStrings(v))

	v = [3]int{1, 2, 3}
	assert.Equal(t, vs, toStrings(v))

	v = []interface{}{1, 2, 1 + 2i}
	assert.Empty(t, toStrings(v))
}
