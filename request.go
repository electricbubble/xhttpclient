package xhttpclient

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

type RequestBuilder interface {
	Context(ctx context.Context) RequestBuilder
	Timeout(timeout time.Duration) RequestBuilder
	Path(elem ...string) RequestBuilder
	Query(query url.Values) RequestBuilder
	QuerySet(key, value string) RequestBuilder
	QueryAdd(key, value string) RequestBuilder
	Header(header http.Header) RequestBuilder
	HeaderSet(key, value string) RequestBuilder
	HeaderAdd(key, value string) RequestBuilder
	HeaderSetBasicAuth(username, password string) RequestBuilder
	Body(body any) RequestBuilder

	build(enc Encoder, opts ...xRequestBuildOption) (req *http.Request, cancel context.CancelFunc, err error)
}

func NewGet() RequestBuilder     { return newRequestBuilder(http.MethodGet) }
func NewHead() RequestBuilder    { return newRequestBuilder(http.MethodHead) }
func NewPost() RequestBuilder    { return newRequestBuilder(http.MethodPost) }
func NewPut() RequestBuilder     { return newRequestBuilder(http.MethodPut) }
func NewPatch() RequestBuilder   { return newRequestBuilder(http.MethodPatch) }
func NewDelete() RequestBuilder  { return newRequestBuilder(http.MethodDelete) }
func NewConnect() RequestBuilder { return newRequestBuilder(http.MethodConnect) }
func NewOptions() RequestBuilder { return newRequestBuilder(http.MethodOptions) }
func NewTrace() RequestBuilder   { return newRequestBuilder(http.MethodTrace) }

func newRequestBuilder(method string) *xRequestBuilder {
	return &xRequestBuilder{
		method: method,
		m:      new(xRequestMutation),
	}
}

var _ RequestBuilder = (*xRequestBuilder)(nil)

type xRequestBuilder struct {
	method  string
	ctx     context.Context
	timeout time.Duration
	baseURL string
	path    []string
	q       xQuery
	m       *xRequestMutation
}

type xRequestBuildOption func(r *xRequestBuilder)

func (r *xRequestBuilder) build(enc Encoder, opts ...xRequestBuildOption) (req *http.Request, cancel context.CancelFunc, err error) {
	for _, opt := range opts {
		opt(r)
	}

	u, err := r.buildURL()
	if err != nil {
		return nil, nil, err
	}

	body, err := enc.Encode(r.m)
	if err != nil {
		return nil, nil, fmt.Errorf("encode: %w", err)
	}

	ctx := cmp.Or(r.ctx, context.Background())
	if r.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	if req, err = http.NewRequestWithContext(ctx, r.method, u.String(), body); err != nil {
		cancel()
		return nil, nil, err
	}

	req.Header = r.m.h.header
	return
}

func (r *xRequestBuilder) buildURL() (u *url.URL, err error) {
	var elems []string
	switch pl := len(r.path); {
	case r.baseURL == "" && pl == 0:
		return nil, errors.New("empty url")

	case r.baseURL == "" && pl == 1:
		u, err = url.Parse(r.path[0])

	case r.baseURL == "" && pl >= 2:
		u, err = url.Parse(r.path[0])
		elems = r.path[1:]

	case pl == 0:
		u, err = url.Parse(r.baseURL)

	default:
		var tmp *url.URL
		if tmp, err = url.Parse(r.path[0]); err != nil {
			break
		}
		if tmp.IsAbs() {
			u = tmp
			elems = r.path[1:]
			break
		}

		u, err = url.Parse(r.baseURL)
		elems = r.path
	}
	if err != nil {
		return nil, err
	}

	var patch *url.URL
	if len(elems) > 0 {
		if patch, err = url.Parse(path.Join(elems...)); err != nil {
			return nil, err
		}
		u = u.JoinPath(patch.EscapedPath())
	}

	u.RawQuery, err = r.q.encode(u, patch)
	return
}

func (r *xRequestBuilder) Context(ctx context.Context) RequestBuilder {
	r.ctx = ctx
	return r
}

func (r *xRequestBuilder) Timeout(timeout time.Duration) RequestBuilder {
	r.timeout = timeout
	return r
}

func (r *xRequestBuilder) Path(elem ...string) RequestBuilder {
	if l := len(elem); cap(r.path) < l {
		r.path = make([]string, 0, l)
	}
	r.path = append(r.path, elem...)
	return r
}

func (r *xRequestBuilder) Query(query url.Values) RequestBuilder {
	r.q.Query(query)
	return r
}

func (r *xRequestBuilder) QuerySet(key, value string) RequestBuilder {
	r.q.QuerySet(key, value)
	return r
}

func (r *xRequestBuilder) QueryAdd(key, value string) RequestBuilder {
	r.q.QueryAdd(key, value)
	return r
}

func (r *xRequestBuilder) Header(header http.Header) RequestBuilder {
	r.m.Header(header)
	return r
}

func (r *xRequestBuilder) HeaderSet(key, value string) RequestBuilder {
	r.m.HeaderSet(key, value)
	return r
}

func (r *xRequestBuilder) HeaderAdd(key, value string) RequestBuilder {
	r.m.HeaderAdd(key, value)
	return r
}

func (r *xRequestBuilder) HeaderSetBasicAuth(username, password string) RequestBuilder {
	r.m.HeaderSetBasicAuth(username, password)
	return r
}

func (r *xRequestBuilder) Body(body any) RequestBuilder {
	r.m.hasBody, r.m.body = true, body
	return r
}
