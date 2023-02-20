package xhttpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

var (
	BodyCodecJSON BodyCodec = &bodyCodecJSON{}

	_bcPoolJSON = sync.Pool{
		New: func() any {
			return new(bodyCodecJSON)
		},
	}

	_ BodyCodec                     = (*bodyCodecJSON)(nil)
	_ BodyCodecResponseIsSuccessful = (*bodyCodecJSON)(nil)
	_ BodyHeaderContentType         = (*bodyCodecJSON)(nil)
	_ BodyHeaderContentLength       = (*bodyCodecJSON)(nil)
	_ BodyHeaderAccept              = (*bodyCodecJSON)(nil)
)

type bodyCodecJSON struct {
	buf *bytes.Buffer

	contentLength int
}

func (bcp *bodyCodecJSON) Get() BodyCodec {
	bc := _bcPoolJSON.Get().(*bodyCodecJSON)
	bc.buf = getBytesBuffer()
	return bc
}

func (bcp *bodyCodecJSON) Put(bc BodyCodec) {
	putBytesBuffer(bc.(*bodyCodecJSON).buf)
	bc.(*bodyCodecJSON).contentLength = 0
	_bcPoolJSON.Put(bc)
}

func (bcp *bodyCodecJSON) Encode(body any) (io.Reader, error) {
	if err := json.NewEncoder(bcp.buf).Encode(body); err != nil {
		return nil, err
	}
	bcp.contentLength = bcp.buf.Len()
	return bcp.buf, nil
}

func (bcp *bodyCodecJSON) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (bcp *bodyCodecJSON) IsSuccessful(resp *http.Response) bool {
	return ResponseIsSuccessfulGTE200LTE299(resp)
}

func (bcp *bodyCodecJSON) ContentType() string {
	return ContentTypeValueJSON
}

func (bcp *bodyCodecJSON) ContentLength() int {
	return bcp.contentLength
}

func (bcp *bodyCodecJSON) Accept() string {
	return ContentTypeValueJSON
}
