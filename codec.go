package xhttpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type (
	RequestMutation interface {
		GetBody() (v any, ok bool)
		GetHeader() http.Header

		Header(header http.Header) RequestMutation
		HeaderSet(key, value string) RequestMutation
		HeaderAdd(key, value string) RequestMutation
		HeaderSetBasicAuth(username, password string) RequestMutation
	}

	ResponseHandle interface {
		StatusCode() int
		Header() http.Header
		Body() io.ReadCloser
		RequestMethod() string
		RequestURL() *url.URL

		BodyReadAll() error
		BodyHasBeenRead() []byte

		Close() error

		Raw() *http.Response
		Cancel()
	}
)

var _ RequestMutation = (*xRequestMutation)(nil)

type xRequestMutation struct {
	h xHeader

	hasBody bool
	body    any
}

func (m *xRequestMutation) GetBody() (v any, ok bool) {
	return m.body, m.hasBody
}

func (m *xRequestMutation) GetHeader() http.Header {
	return m.h.header.Clone()
}

func (m *xRequestMutation) Header(header http.Header) RequestMutation {
	m.h.Header(header)
	return m
}

func (m *xRequestMutation) HeaderSet(key, value string) RequestMutation {
	m.h.HeaderSet(key, value)
	return m
}

func (m *xRequestMutation) HeaderAdd(key, value string) RequestMutation {
	m.h.HeaderAdd(key, value)
	return m
}

func (m *xRequestMutation) HeaderSetBasicAuth(username, password string) RequestMutation {
	m.h.HeaderSetBasicAuth(username, password)
	return m
}

var (
	_                           ResponseHandle = (*xResponseHandle)(nil)
	ErrUnreachableResponseIsNil                = errors.New("unreachable: *http.Response is nil")

	ErrUnhandled = errors.New("unhandled")
)

type xResponseHandle struct {
	raw    *http.Response
	cancel context.CancelFunc

	bodyHasBeenRead []byte
}

func (h *xResponseHandle) StatusCode() int {
	if r := h.Raw(); r != nil {
		return r.StatusCode
	}

	panic(ErrUnreachableResponseIsNil)
}

func (h *xResponseHandle) Header() http.Header {
	if r := h.Raw(); r != nil {
		return r.Header
	}

	panic(ErrUnreachableResponseIsNil)
}

func (h *xResponseHandle) Body() io.ReadCloser {
	if r := h.Raw(); r != nil {
		return r.Body
	}

	panic(ErrUnreachableResponseIsNil)
}

func (h *xResponseHandle) RequestMethod() string {
	r := h.Raw()
	if r == nil {
		return ""
	}

	if r.Request == nil {
		return ""
	}
	return r.Request.Method
}

func (h *xResponseHandle) RequestURL() *url.URL {
	r := h.Raw()
	if r == nil {
		return nil
	}

	if r.Request == nil {
		return nil
	}
	return r.Request.URL
}

func (h *xResponseHandle) BodyReadAll() (err error) {
	h.bodyHasBeenRead, err = io.ReadAll(h.Body())
	return
}

func (h *xResponseHandle) BodyHasBeenRead() []byte {
	return h.bodyHasBeenRead
}

func (h *xResponseHandle) Close() (err error) {
	if body := h.Body(); body != nil {
		_, err1 := io.Copy(io.Discard, body)
		err2 := body.Close()
		err = errors.Join(err1, err2)
	}

	h.Cancel()
	return
}

func (h *xResponseHandle) Raw() *http.Response {
	return h.raw
}

func (h *xResponseHandle) Cancel() {
	h.cancel()
}

type Codec interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(m RequestMutation) (body io.Reader, err error)
}

type Decoder interface {
	Decode(resp ResponseHandle, successV, failureV any) error
}

const (
	HeaderKeyContentType   = "Content-Type"
	HeaderKeyContentLength = "Content-Length"

	HeaderValueJSON = "application/json; charset=utf-8"
)

var (
	CodecNoop              Codec = new(xCodecNoop)
	CodecJSON              Codec = new(xJSON)
	CodecMultipartFormData Codec = new(xMultipartFormData)
)

var (
	_ Encoder = (*EncoderJSON)(nil)
	_ Decoder = (*DecoderJSON)(nil)
)

type xCodecNoop struct{}

func (c *xCodecNoop) Encode(RequestMutation) (body io.Reader, err error) {
	return nil, nil
}

func (c *xCodecNoop) Decode(_ ResponseHandle, _, _ any) error {
	return nil
}
