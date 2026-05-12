package xhttpclient

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"
)

type HTTPClientDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client interface {
	Doer(doer HTTPClientDoer) Client
	RequestTimeout(timeout time.Duration) Client
	BaseURL(baseURL string) Client
	Header(header http.Header) Client
	HeaderSet(key, value string) Client
	HeaderAdd(key, value string) Client
	HeaderSetBasicAuth(username, password string) Client

	Do(rb RequestBuilder, codec Codec, successV, failureV any) (ResponseHandle, error)
}

func NewClient() Client {
	return new(xClient).
		Doer(DefaultHTTPClientDoer)
}

var DefaultHTTPClientDoer HTTPClientDoer = &http.Client{
	Transport: DefaultTransport(),
	Timeout:   0,
}

func DefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		IdleConnTimeout:     45 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: runtime.NumCPU() + 1,
		TLSHandshakeTimeout: 5 * time.Second,
		ForceAttemptHTTP2:   true,
	}
}

var _ Client = (*xClient)(nil)

type xClient struct {
	doer       HTTPClientDoer
	reqTimeout time.Duration
	baseURL    string
	h          xHeader
}

func (c *xClient) applyRequest(r *xRequestBuilder) {
	r.timeout = c.reqTimeout
	r.baseURL = c.baseURL
	r.m.h.apply(c.h)
}

func (c *xClient) Doer(doer HTTPClientDoer) Client {
	c.doer = doer
	return c
}

func (c *xClient) RequestTimeout(timeout time.Duration) Client {
	c.reqTimeout = timeout
	return c
}

func (c *xClient) BaseURL(baseURL string) Client {
	c.baseURL = baseURL
	return c
}

func (c *xClient) Header(header http.Header) Client {
	c.h.Header(header)
	return c
}

func (c *xClient) HeaderSet(key, value string) Client {
	c.h.HeaderSet(key, value)
	return c
}

func (c *xClient) HeaderAdd(key, value string) Client {
	c.h.HeaderAdd(key, value)
	return c
}

func (c *xClient) HeaderSetBasicAuth(username, password string) Client {
	c.h.HeaderSetBasicAuth(username, password)
	return c
}

func (c *xClient) Do(rb RequestBuilder, codec Codec, successV, failureV any) (resp ResponseHandle, err error) {
	if codec == nil {
		panic(errors.New("codec must not be nil"))
	}

	req, cancel, err := rb.build(codec, c.applyRequest)
	if err != nil {
		return nil, err
	}

	raw, err := c.doer.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}

	resp = &xResponseHandle{
		raw:    raw,
		cancel: cancel,
	}
	if err = codec.Decode(resp, successV, failureV); err != nil {
		resp = nil
		cancel()
		return nil, fmt.Errorf("decode: %w", err)
	}
	return
}
