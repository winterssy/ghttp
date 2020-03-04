package ghttp_test

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	neturl "net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/winterssy/ghttp"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/time/rate"
)

func Example_goStyleAPI() {
	client := ghttp.New()

	req, err := ghttp.NewRequest(ghttp.MethodPost, "https://httpbin.org/post")
	if err != nil {
		log.Fatal(err)
	}

	req.SetQuery(ghttp.Params{
		"k1": "v1",
		"k2": "v2",
	})
	req.SetHeaders(ghttp.Headers{
		"k3": "v3",
		"k4": "v4",
	})
	req.SetForm(ghttp.Form{
		"k5": "v5",
		"k6": "v6",
	})

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}

func Example_requestsStyleAPI() {
	client := ghttp.New()

	resp, err := client.Post("https://httpbin.org/post",
		ghttp.WithQuery(ghttp.Params{
			"k1": "v1",
			"k2": "v2",
		}),
		ghttp.WithHeaders(ghttp.Headers{
			"k3": "v3",
			"k4": "v4",
		}),
		ghttp.WithForm(ghttp.Form{
			"k5": "v5",
			"k6": "v6",
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}

func ExampleNoRedirect() {
	client := ghttp.New()
	client.CheckRedirect = ghttp.NoRedirect

	resp, err := client.
		Get("https://httpbin.org/redirect/3")
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(resp.StatusCode)
	// Output:
	// 302
}

func ExampleMaxRedirects() {
	client := ghttp.New()
	client.CheckRedirect = ghttp.MaxRedirects(3)

	resp, err := client.Get("https://httpbin.org/redirect/1")
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(resp.StatusCode)

	resp, err = client.Get("https://httpbin.org/redirect/5")
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(resp.StatusCode)

	// Output:
	// 200
	// 302
}

func ExampleClient_SetProxy() {
	client := ghttp.New()
	client.SetProxy(ghttp.ProxyURL("socks5://127.0.0.1:1080"))

	resp, err := client.Get("https://www.google.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}

func ExampleClient_AddCookies() {
	client := ghttp.New()

	cookies := ghttp.Cookies{
		"n1": "v1",
		"n2": "v2",
	}
	client.AddCookies("https://httpbin.org", cookies.Decode()...)

	resp, err := client.Get("https://httpbin.org/cookies")
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("cookies", "n1"))
	fmt.Println(result.GetString("cookies", "n2"))
	// Output:
	// v1
	// v2
}

func ExampleClient_Cookie() {
	client := ghttp.New()

	_, err := client.Get("https://httpbin.org/cookies/set/uid/10086")
	if err != nil {
		log.Print(err)
		return
	}

	c, err := client.Cookie("https://httpbin.org", "uid")
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(c.Name)
	fmt.Println(c.Value)
	// Output:
	// uid
	// 10086
}

func ExampleClient_DisableTLSVerify() {
	client := ghttp.New()
	client.DisableTLSVerify()

	resp, err := client.Get("https://self-signed.badssl.com")
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(resp.StatusCode)
	// Output:
	// 200
}

func ExampleClient_AddClientCerts() {
	client := ghttp.New()

	cert, err := tls.LoadX509KeyPair("/path/client.cert", "/path/client.key")
	if err != nil {
		log.Fatal(err)
	}

	client.AddClientCerts(cert)
	resp, err := client.Get("https://self-signed.badssl.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}

func ExampleClient_AddRootCerts() {
	pemCerts, err := ioutil.ReadFile("/path/root-ca.pem")
	if err != nil {
		log.Fatal(err)
	}

	client := ghttp.New()
	client.AddRootCerts(pemCerts)

	resp, err := client.Get("https://self-signed.badssl.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}

func ExampleClient_RegisterBeforeRequestCallbacks() {
	withReverseProxy := func(target string) ghttp.RequestHook {
		return func(req *ghttp.Request) error {
			u, err := neturl.Parse(target)
			if err != nil {
				return err
			}

			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.Host = u.Host
			req.SetOrigin(u.Host)
			return nil
		}
	}

	client := ghttp.New()
	client.RegisterBeforeRequestCallbacks(withReverseProxy("https://httpbin.org"))

	resp, err := client.Get("/get")
	if err == nil {
		fmt.Println(resp.StatusCode)
	}
	resp, err = client.Post("/post")
	if err == nil {
		fmt.Println(resp.StatusCode)
	}
	// Output:
	// 200
	// 200
}

func ExampleClient_EnableRateLimiting() {
	client := ghttp.New()
	client.EnableRateLimiting(rate.NewLimiter(1, 10))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.Get("https://www.example.com")
		}()
	}
	wg.Wait()
}

func ExampleClient_SetMaxConcurrency() {
	client := ghttp.New()
	client.SetMaxConcurrency(32)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.Get("https://www.example.com")
		}()
	}
	wg.Wait()
}

func ExampleClient_EnableDebugging() {
	client := ghttp.New()
	client.EnableDebugging(os.Stdout, true)

	_, _ = client.Post("https://httpbin.org/post",
		ghttp.WithForm(ghttp.Form{
			"k1": "v1",
			"k2": "v2",
		}),
	)
}

func ExampleWithQuery() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithQuery(ghttp.Params{
				"k1": "v1",
				"k2": "v2",
			}),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("args", "k1"))
	fmt.Println(result.GetString("args", "k2"))
	// Output:
	// v1
	// v2
}

func ExampleWithHeaders() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithHeaders(ghttp.Headers{
				"k1": "v1",
				"k2": "v2",
			}),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("headers", "K1"))
	fmt.Println(result.GetString("headers", "K2"))
	// Output:
	// v1
	// v2
}

func ExampleWithUserAgent() {
	client := ghttp.New()

	resp, err := client.
		Get("https://httpbin.org/get",
			ghttp.WithUserAgent("ghttp"),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("headers", "User-Agent"))
	// Output:
	// ghttp
}

func ExampleWithReferer() {
	client := ghttp.New()

	resp, err := client.
		Get("https://httpbin.org/get",
			ghttp.WithReferer("https://www.google.com"),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("headers", "Referer"))
	// Output:
	// https://www.google.com
}

func ExampleWithOrigin() {
	client := ghttp.New()

	resp, err := client.
		Get("https://httpbin.org/get",
			ghttp.WithOrigin("https://www.google.com"),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("headers", "Origin"))
	// Output:
	// https://www.google.com
}

func ExampleWithBasicAuth() {
	client := ghttp.New()

	resp, err := client.
		Get("https://httpbin.org/basic-auth/admin/pass",
			ghttp.WithBasicAuth("admin", "pass"),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetBoolean("authenticated"))
	fmt.Println(result.GetString("user"))
	// Output:
	// true
	// admin
}

func ExampleWithBearerToken() {
	client := ghttp.New()

	resp, err := client.
		Get("https://httpbin.org/bearer",
			ghttp.WithBearerToken("ghttp"),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetBoolean("authenticated"))
	fmt.Println(result.GetString("token"))
	// Output:
	// true
	// ghttp
}

func ExampleWithCookies() {
	client := ghttp.New()

	resp, err := client.
		Get("https://httpbin.org/cookies",
			ghttp.WithCookies(ghttp.Cookies{
				"n1": "v1",
				"n2": "v2",
			}),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("cookies", "n1"))
	fmt.Println(result.GetString("cookies", "n2"))
	// Output:
	// v1
	// v2
}

func ExampleWithBody() {
	client := ghttp.New()

	formData := ghttp.
		NewMultipart(ghttp.Files{
			"file1": ghttp.MustOpen("./testdata/testfile1.txt"),
			"file2": ghttp.MustOpen("./testdata/testfile2.txt"),
		}).
		WithForm(ghttp.Form{
			"k1": "v1",
			"k2": "v2",
		})

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithBody(formData),
			ghttp.WithContentType(formData.ContentType()),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("files", "file1"))
	fmt.Println(result.GetString("files", "file2"))
	fmt.Println(result.GetString("form", "k1"))
	fmt.Println(result.GetString("form", "k2"))
	// Output:
	// testfile1.txt
	// testfile2.txt
	// v1
	// v2
}

func ExampleWithContent() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithContent([]byte("hello world")),
			ghttp.WithContentType("text/plain; charset=utf-8"),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("data"))
	// Output:
	// hello world
}

func ExampleWithText() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithText("hello world"),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("data"))
	// Output:
	// hello world
}

func ExampleWithForm() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithForm(ghttp.Form{
				"k1": "v1",
				"k2": "v2",
			}),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("form", "k1"))
	fmt.Println(result.GetString("form", "k2"))
	// Output:
	// v1
	// v2
}

func ExampleWithJSON() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithJSON(map[string]interface{}{
				"msg": "hello world",
				"num": 2019,
			}),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("json", "msg"))
	fmt.Println(result.GetNumber("json", "num"))
	// Output:
	// hello world
	// 2019
}

func ExampleWithFiles() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithFiles(ghttp.Files{
				"file1": ghttp.MustOpen("./testdata/testfile1.txt"),
				"file2": ghttp.MustOpen("./testdata/testfile2.txt"),
			}),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("files", "file1"))
	fmt.Println(result.GetString("files", "file2"))
	// Output:
	// testfile1.txt
	// testfile2.txt
}

func ExampleWithContext() {
	client := ghttp.New()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := client.
		Get("https://httpbin.org/delay/10",
			ghttp.WithContext(ctx),
		)
	fmt.Println(err)
	// Output:
	// Get https://httpbin.org/delay/10: context deadline exceeded
}

func ExampleEnableRetry() {
	client := ghttp.New()

	resp, err := client.Post("https://api.example.com/login",
		ghttp.WithBasicAuth("user", "p@ssw$"),
		ghttp.EnableRetry(
			ghttp.WithRetryMaxAttempts(5),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}

func ExampleEnableClientTrace() {
	client := ghttp.New()

	resp, _ := client.
		Get("https://httpbin.org/get",
			ghttp.EnableClientTrace(),
		)
	fmt.Printf("%+v\n", resp.TraceInfo())
}

func ExampleResponse_Text() {
	client := ghttp.New()

	resp, err := client.Get("https://www.example.com")
	if err != nil {
		log.Fatal(err)
	}
	s, err := resp.Text()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)

	resp, err = client.Get("https://www.example.cn")
	if err != nil {
		log.Fatal(err)
	}
	s, err = resp.Text(simplifiedchinese.GBK)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)
}

func ExampleResponse_H() {
	client := ghttp.New()

	resp, err := client.
		Post("https://httpbin.org/post",
			ghttp.WithQuery(ghttp.Params{
				"k1": "v1",
				"k2": "v2",
			}),
			ghttp.WithHeaders(ghttp.Headers{
				"k3": "v3",
				"k4": "v4",
			}),
			ghttp.WithForm(ghttp.Form{
				"k5": "v5",
				"k6": "v6",
			}),
		)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := resp.H()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(result.GetString("args", "k1"))
	fmt.Println(result.GetString("args", "k2"))
	fmt.Println(result.GetString("headers", "K3"))
	fmt.Println(result.GetString("headers", "K4"))
	fmt.Println(result.GetString("form", "k5"))
	fmt.Println(result.GetString("form", "k6"))
	// Output:
	// v1
	// v2
	// v3
	// v4
	// v5
	// v6
}

const picData = `
iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6
JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAACfFBMVEUAAABVVVU3TlI4T1Q4T1Q4TlQ4
TlQ4TlQ4TlQ3TlQ2T1QAgIA5T1Q4TlU4TlUAAAA5TVM4TlM4TlM4TVMAAAA4TlQ4TlQ4TlQ4TVQ5TFQ4
TlQ4T1Q5TlU4TlQ4T1U3TlU4TVM4T1Q4TlQ4TlU4TlU5UFc5TlQ5Ulc4TlQ5T1Q3T1U4TlQ4TlQ4TlM4
TlQ4TlQ7V19PhJJbobRnvdRtzedx1vF03vpHcn5boLJNfow/X2hfrMB03Pl24f503Pg/X2c7VV1VkqNX
mao+XWVAYmt13flis8lOg5J52/W07/7d9//l+f/S9f+g6v534f6j6/7H8//M9P+z7v6A4/5z2fVUkaFu
z+lz2/dw0u1pw9tJdoJz2PNTjZ1Sjp1z2vaE4/3q+v/////L9P+B4/7g+P/2/f+a6f524P5nvtZRiZhZ
n7Fmu9J23/x24P3Z9//4/f+G5f5UkaJOgZBsyeKX6P719fVfXFsqJiV5d3bx+/6T5/7S0tFAPTsxLiyy
sLDC8v9owNhFbXhjtcqYlpYXExHBwMCH5f6x7v5TUE8kIB7c9Ptv0OuamJfDwsFUUVAlIR/d9fxv0OqV
6P729vZjYV8uKih/fXvv+/6R5/7V1NREQUA2MjG1tLPB8f/e+P92l6A0R0w1S1GarbP3/f+F5P6E5P7o
+v////69yswjIB8qKCfX3t30/f+Y6P574v7j+f/P9f+1qaDDjHSFYlNJOTNMPDWQalmytLDD8v/J8/+v
7v5/4/553PjAjXa3j3x13vtgmql8k5S3n5TBpprCp5uwnJN6k5Zfn7Bz3PhrzObq8PH//PvX5elqyuTN
4ea82uJlwdpcsMZoxt8AAAA/l2hnAAAAL3RSTlMAA0GSvN7u8eOPPQK5ymkCWd/cVgFbw8BJQ/zj3P5X
wdbg9bvQ/lX0+GevsdrZ88Z1oBsAAAABYktHRACIBR1IAAAAB3RJTUUH5AIcDgA5dJXtcQAAAZ9JREFU
OMtjYCAWMDIxs7Cysenrs7NxsHBycSNkuHhYefn4BQT1UYCQsIiomKC4BAODpBSIL62vb2BoZGxiamZm
amJsZGigry8DEpeVY5DXN7ew1Ne3sraxRQJ21vb6Do5OzvoKDIouZq62bu4enl7ePr5gST//gMCg4JDQ
sPCISH0lBuWo6JjYuPiERBBI8rO1TU4BsVLT0jMys0L0VRgkVbMjcnLBgkAQZGubB2HlF9jahhfqyzEw
qKkXFZeAhErLyisq/apAzOqa2rp624ZGDU2QP7X0mzxBos0tLS2tbe0gZkdLS2eXrZu+NjggdPS7g0Ci
PUAFvQVgBX0tLf0TbCfq64IV6OnbTQKJTp4yddp0vxkg5sxZs+fMtbXT1wMr0Ne3tZ0HdeT8BQsXQViL
l9ja6uvDFSxdBhJcvmLlypWrVoOYa9baoiiwXdeet37Dxk2bt2zdtmn7jp27dtuiKQCBPXs3gcE+WICj
K9h/4OChw0eOHjuOS4HtiZOnTp8+c/YETgW2tudOnz6P4GFRcOHixUt4FaCCUQV0VQAAe4oDXnKi80AA
AAAldEVYdGRhdGU6Y3JlYXRlADIwMjAtMDItMjhUMTQ6MDA6NTcrMDA6MDAP9tSBAAAAJXRFWHRkYXRl
Om1vZGlmeQAyMDIwLTAyLTI4VDE0OjAwOjU3KzAwOjAwfqtsPQAAAABJRU5ErkJggg==
`

func ExampleFile_WithFilename() {
	file := ghttp.FileFromReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(picData))).WithFilename("image.jpg")
	_ = ghttp.Files{
		"field": file,
	}
	// Content-Disposition: form-data; name="field"; filename="image.jpg"
}

func ExampleFile_WithMIME() {
	_ = ghttp.MustOpen("/path/image.png").WithMIME("image/png")
	// Content-Type: image/png
}
