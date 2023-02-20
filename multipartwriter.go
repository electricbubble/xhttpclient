package xhttpclient

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type XMultipartWriter struct {
	boundary string
	ops      []func(*multipart.Writer) error
}

var _xMultipartWriterPool = sync.Pool{
	New: func() any {
		return &XMultipartWriter{
			ops: make([]func(writer *multipart.Writer) error, 0, 2),
		}
	},
}

func NewMultipartWriter() *XMultipartWriter {
	xmw := _xMultipartWriterPool.Get().(*XMultipartWriter)
	return xmw.SetBoundary(randomBoundary())
}

func (xmw *XMultipartWriter) Boundary() string {
	return xmw.boundary
}

func (xmw *XMultipartWriter) SetBoundary(boundary string) *XMultipartWriter {
	if (&multipart.Writer{}).SetBoundary(boundary) == nil {
		xmw.boundary = boundary
	}

	xmw.ops = append(xmw.ops, func(raw *multipart.Writer) error {
		return raw.SetBoundary(boundary)
	})
	return xmw
}

// FormDataContentType returns the Content-Type for an HTTP
// multipart/form-data with this Writer's Boundary.
//
// Copied from mime/multipart/writer.go
func (xmw *XMultipartWriter) FormDataContentType() string {
	b := xmw.boundary
	// We must quote the boundary if it contains any of the
	// tspecials characters defined by RFC 2045, or space.
	if strings.ContainsAny(b, `()<>@,;:\"/[]?= `) {
		b = `"` + b + `"`
	}
	return "multipart/form-data; boundary=" + b
}

func (xmw *XMultipartWriter) WriteWithHeader(header textproto.MIMEHeader, r io.Reader) *XMultipartWriter {
	xmw.ops = append(xmw.ops, func(raw *multipart.Writer) (err error) {
		var pw io.Writer
		if pw, err = raw.CreatePart(header); err != nil {
			return err
		}
		_, err = io.Copy(pw, r)
		return err
	})
	return xmw
}

func (xmw *XMultipartWriter) WriteWithFile(fieldname, filename string) *XMultipartWriter {
	xmw.ops = append(xmw.ops, func(raw *multipart.Writer) (err error) {
		var pw io.Writer
		if pw, err = raw.CreateFormFile(fieldname, filepath.Base(filename)); err != nil {
			return err
		}

		var file *os.File
		if file, err = os.Open(filename); err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(pw, file)
		return err
	})

	return xmw
}

func (xmw *XMultipartWriter) WriteWithField(fieldname string, r io.Reader) *XMultipartWriter {
	xmw.ops = append(xmw.ops, func(raw *multipart.Writer) (err error) {
		var pw io.Writer
		if pw, err = raw.CreateFormField(fieldname); err != nil {
			return err
		}

		_, err = io.Copy(pw, r)
		return err
	})
	return xmw
}

func (xmw *XMultipartWriter) WriteWithFieldValue(fieldname, value string) *XMultipartWriter {
	xmw.ops = append(xmw.ops, func(raw *multipart.Writer) (err error) {
		return raw.WriteField(fieldname, value)
	})

	return xmw
}

func (xmw *XMultipartWriter) do(buf *bytes.Buffer) error {
	raw := multipart.NewWriter(buf)
	for _, op := range xmw.ops {
		if err := op(raw); err != nil {
			return err
		}
	}

	return raw.Close()
}

func (xmw *XMultipartWriter) free() {
	xmw.boundary = ""
	xmw.ops = nil
	_xMultipartWriterPool.Put(xmw)
}
