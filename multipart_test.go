package ghttp

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFile_Read(t *testing.T) {
	const (
		msg = "hello world"
	)

	file := FileFromReader(strings.NewReader(msg)).WithFilename("testfile.txt")
	n, err := io.Copy(ioutil.Discard, file)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(len(msg)), n)
	}
}

func TestOpen(t *testing.T) {
	const (
		fileExist    = "./testdata/testfile1.txt"
		fileNotExist = "./testdata/file_not_exist.txt"
	)

	f, err := Open(fileExist)
	if assert.NoError(t, err) {
		assert.NoError(t, f.Close())
	}

	_, err = Open(fileNotExist)
	assert.Error(t, err)
}

func TestMustOpen(t *testing.T) {
	const (
		fileExist    = "./testdata/testfile1.txt"
		fileNotExist = "./testdata/file_not_exist.txt"
		mime         = "text/plain; charset=utf-8"
	)

	var f *File
	if assert.NotPanics(t, func() {
		f = MustOpen(fileExist).WithMIME(mime)
	}) {
		assert.Equal(t, "testfile1.txt", f.filename)
		assert.Equal(t, mime, f.mime)
	}

	assert.Panics(t, func() {
		f = MustOpen(fileNotExist)
	})
}
