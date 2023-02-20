package xhttpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"time"
)

type XClient struct {
	baseURL string
	header  http.Header

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

func (xc *XClient) WithTimeout(d time.Duration) *XClient {
	xc.doer.Timeout = d
	return xc
}

func (xc *XClient) WithJar(jar *cookiejar.Jar) *XClient {
	xc.doer.Jar = jar
	return xc
}

func (xc *XClient) WithCheckRedirect(fn func(req *http.Request, via []*http.Request) error) *XClient {
	xc.doer.CheckRedirect = fn
	return xc
}

func (xc *XClient) WithTransport(transport http.RoundTripper) *XClient {
	xc.doer.Transport = transport
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
	var (
		req *http.Request
		bc  = bodyCodec.Get()
	)
	defer bodyCodec.Put(bc)

	req, resp, err = xc.do(bc, xReq)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

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
		if err := decodeWrong(bc, bytes.NewBuffer(respBody), wrongV); err != nil {
			return resp, respBody, decodeErrorF(resp.StatusCode)
		}
	case isSuccessful(bc, resp):
		if err := bc.Decode(bytes.NewBuffer(respBody), successV); err != nil {
			return resp, respBody, decodeErrorF(resp.StatusCode)
		}
	default:
		if err := bc.Decode(bytes.NewBuffer(respBody), wrongV); err != nil {
			return resp, respBody, decodeErrorF(resp.StatusCode)
		}
	}

	return
}

func (xc *XClient) DoWithRaw(xReq *XRequestBuilder) (req *http.Request, resp *http.Response, err error) {
	codec := xc.bodyCodecPool.Get()
	defer xc.bodyCodecPool.Put(codec)
	req, resp, err = xc.do(codec, xReq)
	return
}

func (xc *XClient) do(bc BodyCodec, xReq *XRequestBuilder) (req *http.Request, resp *http.Response, err error) {
	xc.initXReq(xReq)
	if req, err = xReq.build(bc); err != nil {
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

func decodeErrorF(statusCode int) error {
	return fmt.Errorf("unexpected error (HTTP status: %s[%d])", http.StatusText(statusCode), statusCode)
}
