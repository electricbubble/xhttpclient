package xhttpclient

import (
	"io"
	"net/http"
	"net/textproto"
)

type testTypeBase struct {
	url    string
	header http.Header
	codec  BodyCodec
}

func testTypeBaseEmpty() testTypeBase {
	return testTypeBase{
		url:    "",
		header: nil,
		codec:  &testFakeBodyCodecJSON{fnEncode: func(v any) (io.Reader, error) { return nil, nil }},
	}
}

func initTestXReq(base testTypeBase, xReq *XRequestBuilder) {
	xReq.baseURL = base.url
	if base.header != nil && xReq.header != nil {
		for cliHdrKey, vv := range base.header {
			k := textproto.CanonicalMIMEHeaderKey(cliHdrKey)
			if _, ok := xReq.header[k]; ok {
				continue
			}
			xReq.header[k] = vv
		}
	}
}
