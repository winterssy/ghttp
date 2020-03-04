package ghttp

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/winterssy/gjson"
)

type (
	// H is an alias of gjson.Object.
	// Visit https://github.com/winterssy/gjson for more details.
	H = gjson.Object
)

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func toString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return b2s(v)
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case error:
		return v.Error()
	case interface {
		String() string
	}:
		return v.String()
	}

	panic(fmt.Errorf("ghttp: can't cast %#v of type %[1]T to string", v))
}

func toStrings(v interface{}) (vs []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
			vs = nil // ignore this field
		}
	}()

	switch v := v.(type) {
	case []string:
		vs = make([]string, len(v))
		copy(vs, v)
	case string, []byte,
		bool,
		float32, float64,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		error,
		interface {
			String() string
		}:
		vs = []string{toString(v)}
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			n := rv.Len()
			vs = make([]string, 0, n)
			for i := 0; i < n; i++ {
				vs = append(vs, toString(rv.Index(i).Interface()))
			}
		}
	}

	return
}

func toReadCloser(r io.Reader) io.ReadCloser {
	rc, ok := r.(io.ReadCloser)
	if !ok && r != nil {
		rc = ioutil.NopCloser(r)
	}
	return rc
}

// Report whether an HTTP request or response body is empty.
func bodyEmpty(body io.ReadCloser) bool {
	return body == nil || body == http.NoBody
}

func drainBody(body io.ReadCloser, w io.Writer) (err error) {
	defer body.Close()
	_, err = io.Copy(w, body)
	return
}

func findCookie(name string, cookies []*http.Cookie) (*http.Cookie, error) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}

	return nil, http.ErrNoCookie
}

// Return value if nonempty, def otherwise.
func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}
