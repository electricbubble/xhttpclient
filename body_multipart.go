package xhttpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
)

var (
	BodyCodecMultipart BodyCodec = &bodyCodecMultipart{}

	_bcPoolMultipart = sync.Pool{
		New: func() any {
			return new(bodyCodecMultipart)
		},
	}

	_ BodyCodec                     = (*bodyCodecMultipart)(nil)
	_ BodyCodecResponseIsSuccessful = (*bodyCodecMultipart)(nil)
	_ BodyHeaderContentType         = (*bodyCodecMultipart)(nil)
	_ BodyHeaderContentLength       = (*bodyCodecMultipart)(nil)
	_ BodyHeaderAccept              = (*bodyCodecMultipart)(nil)
)

type bodyCodecMultipart struct {
	buf *bytes.Buffer

	contentType   string
	contentLength int
}

func (bcp *bodyCodecMultipart) Get() BodyCodec {
	bc := _bcPoolMultipart.Get().(*bodyCodecMultipart)
	bc.buf = getBytesBuffer()
	return bc
}

func (bcp *bodyCodecMultipart) Put(bc BodyCodec) {
	putBytesBuffer(bc.(*bodyCodecMultipart).buf)
	bc.(*bodyCodecMultipart).contentType = ""
	bc.(*bodyCodecMultipart).contentLength = 0
	_bcPoolMultipart.Put(bc)
}

func (bcp *bodyCodecMultipart) Encode(body any) (io.Reader, error) {
	xmw, ok := body.(*XMultipartWriter)
	if !ok {
		return nil, fmt.Errorf(
			"expected body type '%s', got unconvertible value type '%T'",
			reflect.TypeOf(&XMultipartWriter{}).Name(), body,
		)
	}
	defer xmw.free()

	if err := xmw.do(bcp.buf); err != nil {
		return bcp.buf, err
	}

	bcp.contentType = xmw.FormDataContentType()
	bcp.contentLength = bcp.buf.Len()
	return bcp.buf, nil
}

func (bcp *bodyCodecMultipart) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (bcp *bodyCodecMultipart) IsSuccessful(resp *http.Response) bool {
	return ResponseIsSuccessfulGTE200LTE299(resp)
}

func (bcp *bodyCodecMultipart) ContentType() string {
	return bcp.contentType
}

func (bcp *bodyCodecMultipart) ContentLength() int {
	return bcp.contentLength
}

func (bcp *bodyCodecMultipart) Accept() string {
	return ContentTypeValueJSON
}
