package xhttpclient

import (
	"bytes"
	"io"
	"mime/multipart"
	"strconv"
)

type xMultipartFormData struct {
	DecoderJSON
}

func (c *xMultipartFormData) Encode(m RequestMutation) (body io.Reader, err error) {
	v, ok := m.GetBody()
	if !ok {
		return nil, nil
	}

	fns := v.([]func(*multipart.Writer) error)

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	for _, fn := range fns {
		if err = fn(w); err != nil {
			return nil, err
		}
	}
	if err = w.Close(); err != nil {
		return nil, err
	}

	m.HeaderSet(HeaderKeyContentType, w.FormDataContentType())
	m.HeaderSet(HeaderKeyContentLength, strconv.Itoa(buf.Len()))

	body = buf
	return
}
