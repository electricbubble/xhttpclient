package xhttpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"time"
)

type XClient struct {
	baseURL    string
	header     http.Header
	reqTimeout time.Duration

	doer          *http.Client
	bodyCodecPool BodyCodec
}

func NewClient() *XClient {
	return (&XClient{
		header: make(http.Header),
		doer:   DefaultClient(),
	}).
		WithBodyCodecJSON()
}

func (xc *XClient) Clone() *XClient {
	return &XClient{
		baseURL: xc.baseURL,
		header:  xc.header.Clone(),
		doer:    xc.doer,
	}
}

func (xc *XClient) WithBodyCodecJSON() *XClient {
	return xc.WithBodyCodec(BodyCodecJSON)
}

func (xc *XClient) WithBodyCodec(bodyCodec BodyCodec) *XClient {
	xc.bodyCodecPool = bodyCodec
	return xc
}

func (xc *XClient) WithRequestTimeout(d time.Duration) *XClient {
	xc.reqTimeout = d
	return xc
}

func (xc *XClient) WithClient(c *http.Client) *XClient {
	xc.doer = c
	return xc
}

func (xc *XClient) BaseURL(baseURL string) *XClient {
	xc.baseURL = baseURL
	return xc
}

func (xc *XClient) Header(header http.Header) *XClient {
	xc.header = header.Clone()
	return xc
}

func (xc *XClient) SetHeader(key, value string) *XClient {
	if xc.header == nil {
		xc.header = make(http.Header)
	}
	xc.header.Set(key, value)
	return xc
}

func (xc *XClient) AddHeader(key, value string) *XClient {
	if xc.header == nil {
		xc.header = make(http.Header)
	}
	xc.header.Add(key, value)
	return xc
}

func (xc *XClient) SetBasicAuth(username, password string) *XClient {
	return xc.SetHeader("Authorization", "Basic "+basicAuth(username, password))
}

func (xc *XClient) Do(successV, wrongV any, xReq *XRequestBuilder) (resp *http.Response, respBody []byte, err error) {
	return xc.DoOnceWithBodyCodec(xc.bodyCodecPool, successV, wrongV, xReq)
}

func (xc *XClient) DoOnceWithBodyCodec(bodyCodec BodyCodec, successV, wrongV any, xReq *XRequestBuilder) (resp *http.Response, respBody []byte, err error) {
	if successV == nil {
		return nil, nil, errors.New("'successV' must not be nil")
	}

	var (
		req    *http.Request
		cancel context.CancelFunc
		bc     = bodyCodec.Get()
	)
	defer bodyCodec.Put(bc)

	req, resp, cancel, err = xc.do(bc, xReq)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	defer wrapCancelAndCloseRespBody(cancel, resp)()

	if bc, ok := bc.(BodyCodecOnReceive); ok {
		bc.OnReceive(req, resp)
	}

	if resp.StatusCode == http.StatusNoContent {
		return resp, nil, nil
	}

	if respBody, err = io.ReadAll(resp.Body); err != nil {
		return resp, nil, fmt.Errorf("copy response body: %w", err)
	}
	switch {
	case isWrong(bc, resp):
		if wrongV == nil {
			return resp, respBody, wrapDecodeError(nil, resp)
		}
		if err := decodeWrong(bc, bytes.NewBuffer(respBody), wrongV); err != nil {
			return resp, respBody, wrapDecodeError(err, resp)
		}
	case isSuccessful(bc, resp):
		if err := bc.Decode(bytes.NewBuffer(respBody), successV); err != nil {
			return resp, respBody, wrapDecodeError(err, resp)
		}
	default:
		if wrongV == nil {
			return resp, respBody, wrapDecodeError(nil, resp)
		}
		if err := bc.Decode(bytes.NewBuffer(respBody), wrongV); err != nil {
			return resp, respBody, wrapDecodeError(err, resp)
		}
	}

	return
}

func (xc *XClient) DoWithRaw(xReq *XRequestBuilder) (req *http.Request, resp *http.Response, cancel context.CancelFunc, err error) {
	codec := xc.bodyCodecPool.Get()
	defer xc.bodyCodecPool.Put(codec)
	req, resp, cancel, err = xc.do(codec, xReq)
	return req, resp, wrapCancelAndCloseRespBody(cancel, resp), err
}

func (xc *XClient) do(bc BodyCodec, xReq *XRequestBuilder) (req *http.Request, resp *http.Response, cancel context.CancelFunc, err error) {
	xc.initXReq(xReq)
	if req, cancel, err = xReq.build(bc); err != nil {
		return
	}

	if bc, ok := bc.(BodyCodecOnSend); ok {
		bc.OnSend(req)
	}

	if resp, err = xc.doer.Do(req); err != nil {
		return
	}

	return
}

func (xc *XClient) initXReq(xReq *XRequestBuilder) {
	xReq.baseURL = xc.baseURL
	if xReq.timeout <= 0 {
		xReq.timeout = xc.reqTimeout
	}

	cliHdrLen, reqHdrLen := len(xc.header), len(xReq.header)
	switch {
	case cliHdrLen == 0 && reqHdrLen == 0 || cliHdrLen == 0 && reqHdrLen != 0:
		// no-op
	case cliHdrLen != 0 && reqHdrLen == 0:
		xReq.header = xc.header.Clone()
	default: // cliHdrLen != 0 && reqHdrLen != 0
		for cliHdrKey, vv := range xc.header {
			k := textproto.CanonicalMIMEHeaderKey(cliHdrKey)
			if _, ok := xReq.header[k]; ok {
				continue
			}
			xReq.header[k] = vv
		}
	}
}

func isWrong(bc BodyCodec, resp *http.Response) bool {
	condition, ok := bc.(BodyCodecResponseIsWrong)
	if !ok {
		return false
	}

	return condition.IsWrong(resp)
}

func decodeWrong(bc BodyCodec, r io.Reader, wrongV any) error {
	bcDW, ok := bc.(BodyCodecDecodeWrong)
	if !ok {
		return bc.Decode(r, wrongV)
	}

	return bcDW.DecodeWrong(r, wrongV)
}

func isSuccessful(bc BodyCodec, resp *http.Response) bool {
	if condition, ok := bc.(BodyCodecResponseIsSuccessful); ok {
		return condition.IsSuccessful(resp)
	} else {
		return ResponseIsSuccessfulGTE200LTE299(resp)
	}
}

func wrapDecodeError(err error, resp *http.Response) error {
	if err != nil {
		return fmt.Errorf("unexpected error: URL: %s (%d %s): %w", resp.Request.URL, resp.StatusCode, http.StatusText(resp.StatusCode), err)
	}
	return fmt.Errorf("unexpected error: URL: %s (%d %s)", resp.Request.URL, resp.StatusCode, http.StatusText(resp.StatusCode))
}
