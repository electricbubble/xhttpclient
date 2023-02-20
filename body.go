package xhttpclient

import (
	"io"
	"net/http"
)

const (
	ContentTypeValueJSON           = "application/json; charset=utf-8"
	ContentTypeValueFormUrlencoded = "application/x-www-form-urlencoded"
)

type BodyCodec interface {
	Get() BodyCodec
	Put(bc BodyCodec)

	Encode(body any) (io.Reader, error)
	Decode(r io.Reader, v any) error
}

type (
	BodyCodecOnSend interface {
		OnSend(req *http.Request)
	}
	BodyCodecOnReceive interface {
		OnReceive(req *http.Request, resp *http.Response)
	}

	BodyCodecResponseIsSuccessful interface {
		IsSuccessful(resp *http.Response) bool
	}
	BodyCodecResponseIsWrong interface {
		IsWrong(resp *http.Response) bool
	}

	BodyCodecDecodeWrong interface {
		DecodeWrong(r io.Reader, v any) error
	}
)

type (
	BodyHeaderContentLength interface {
		ContentLength() int
	}

	BodyHeaderContentType interface {
		ContentType() string
	}

	BodyHeaderContentEncoding interface {
		ContentEncoding() string
	}
)

type (
	BodyHeaderAccept interface {
		Accept() string
	}

	BodyHeaderAcceptEncoding interface {
		AcceptEncoding() string
	}
)

func ResponseIsSuccessfulGTE200LTE299(resp *http.Response) bool {
	code := resp.StatusCode
	return 200 <= code && code <= 299
}

func BodyCodecResponseIsWrongGTE400(resp *http.Response) bool {
	return resp.StatusCode >= 400
}
