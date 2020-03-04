package ghttp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type (
	// FormData is a multipart container that uses io.Pipe to reduce memory used
	// while uploading files.
	FormData struct {
		files Files
		form  Form
		pr    *io.PipeReader
		pw    *io.PipeWriter
		mw    *multipart.Writer
		once  sync.Once
	}

	// Files maps a string key to a *File type value, used for files of multipart payload.
	Files map[string]*File

	// File is a struct defines a file of a multipart section to upload.
	File struct {
		body     io.ReadCloser
		filename string
		mime     string
	}
)

// NewMultipart returns a new multipart container.
func NewMultipart(files Files) *FormData {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	return &FormData{
		mw:    mw,
		pr:    pr,
		pw:    pw,
		files: files,
	}
}

// WithForm specifies form for fd.
// If you only want to send form payload, use Request.SetForm or ghttp.WithForm instead.
func (fd *FormData) WithForm(form Form) *FormData {
	fd.form = form
	return fd
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (fd *FormData) writeFiles() {
	const (
		fileFormat      = `form-data; name="%s"; filename="%s"`
		unknownFilename = "???"
	)
	var (
		part io.Writer
		err  error
	)
	for k, v := range fd.files {
		filename := valueOrDefault(v.filename, unknownFilename)

		r := bufio.NewReader(v)
		mime := v.mime
		if mime == "" {
			data, _ := r.Peek(512)
			mime = http.DetectContentType(data)
		}

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(fileFormat, escapeQuotes(k), escapeQuotes(filename)))
		h.Set("Content-Type", mime)
		part, _ = fd.mw.CreatePart(h)
		_, err = io.Copy(part, r)
		if err != nil {
			log.Printf("ghttp: can't bind multipart section (%s=@%s): %s", k, filename, err.Error())
		}
		v.Close()
	}
}

func (fd *FormData) writeForm() {
	for k, vs := range fd.form.Decode() {
		for _, v := range vs {
			fd.mw.WriteField(k, v)
		}
	}
}

// Read implements io.Reader interface.
func (fd *FormData) Read(b []byte) (int, error) {
	fd.once.Do(func() {
		go func() {
			defer fd.pw.Close()
			defer fd.mw.Close() // must close the multipart writer first!
			fd.writeFiles()
			if len(fd.form) > 0 {
				fd.writeForm()
			}
		}()
	})
	return fd.pr.Read(b)
}

// ContentType returns the Content-Type for an HTTP
// multipart/form-data with this multipart Container's Boundary.
func (fd *FormData) ContentType() string {
	return fd.mw.FormDataContentType()
}

// WithFilename specifies f's filename.
func (f *File) WithFilename(filename string) *File {
	f.filename = filename
	return f
}

// WithMIME specifies f's Content-Type.
// By default ghttp detects automatically using http.DetectContentType.
func (f *File) WithMIME(mime string) *File {
	f.mime = mime
	return f
}

// Read implements io.Reader interface.
func (f *File) Read(b []byte) (int, error) {
	return f.body.Read(b)
}

// Close implements io.Closer interface.
func (f *File) Close() error {
	return f.body.Close()
}

// FileFromReader constructors a new File from a reader.
func FileFromReader(body io.Reader) *File {
	return &File{body: toReadCloser(body)}
}

// Open opens the named file and returns a File with filename specified.
func Open(filename string) (*File, error) {
	body, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return FileFromReader(body).WithFilename(filepath.Base(filename)), nil
}

// MustOpen is like Open, but if there is an error, it will panic.
func MustOpen(filename string) *File {
	file, err := Open(filename)
	if err != nil {
		panic(err)
	}

	return file
}
