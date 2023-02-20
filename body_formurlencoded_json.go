package xhttpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"reflect"
	"sort"
	"sync"
)

var (
	BodyCodecFormUrlencodedAndJSON BodyCodec = &bodyCodecFormUrlencodedAndJSON{}

	_bcPoolFormUrlencodedAndJSON = sync.Pool{
		New: func() any {
			return new(bodyCodecFormUrlencodedAndJSON)
		},
	}

	_ BodyCodec                     = (*bodyCodecFormUrlencodedAndJSON)(nil)
	_ BodyCodecResponseIsSuccessful = (*bodyCodecFormUrlencodedAndJSON)(nil)
	_ BodyHeaderContentType         = (*bodyCodecFormUrlencodedAndJSON)(nil)
	_ BodyHeaderContentLength       = (*bodyCodecFormUrlencodedAndJSON)(nil)
	_ BodyHeaderAccept              = (*bodyCodecFormUrlencodedAndJSON)(nil)
)

type bodyCodecFormUrlencodedAndJSON struct {
	buf *bytes.Buffer

	contentLength int
}

func (bcp *bodyCodecFormUrlencodedAndJSON) Get() BodyCodec {
	bc := _bcPoolFormUrlencodedAndJSON.Get().(*bodyCodecFormUrlencodedAndJSON)
	bc.buf = getBytesBuffer()
	return bc
}

func (bcp *bodyCodecFormUrlencodedAndJSON) Put(bc BodyCodec) {
	putBytesBuffer(bc.(*bodyCodecFormUrlencodedAndJSON).buf)
	bc.(*bodyCodecFormUrlencodedAndJSON).contentLength = 0
	_bcPoolFormUrlencodedAndJSON.Put(bc)
}

func (bcp *bodyCodecFormUrlencodedAndJSON) Encode(body any) (io.Reader, error) {
	if body == nil {
		return bcp.buf, nil
	}

	var v urlpkg.Values
	switch tv := body.(type) {
	case urlpkg.Values:
		v = tv
	case *urlpkg.Values:
		v = *tv
	default:
		return nil, fmt.Errorf(
			"expected body type '%s', got unconvertible value type '%T'",
			reflect.TypeOf(urlpkg.Values{}).Name(), body,
		)
	}

	// Copied from net/url/url.go
	// _ = (urlpkg.Values{}).Encode
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := urlpkg.QueryEscape(k)
		for _, v := range vs {
			if bcp.buf.Len() > 0 {
				bcp.buf.WriteByte('&')
			}
			bcp.buf.WriteString(keyEscaped)
			bcp.buf.WriteByte('=')
			bcp.buf.WriteString(urlpkg.QueryEscape(v))
		}
	}

	bcp.contentLength = bcp.buf.Len()
	return bcp.buf, nil
}

func (bcp *bodyCodecFormUrlencodedAndJSON) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (bcp *bodyCodecFormUrlencodedAndJSON) IsSuccessful(resp *http.Response) bool {
	return ResponseIsSuccessfulGTE200LTE299(resp)
}

func (bcp *bodyCodecFormUrlencodedAndJSON) ContentType() string {
	return ContentTypeValueFormUrlencoded
}

func (bcp *bodyCodecFormUrlencodedAndJSON) ContentLength() int {
	return bcp.contentLength
}

func (bcp *bodyCodecFormUrlencodedAndJSON) Accept() string {
	return ContentTypeValueJSON
}
