package xhttpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type XRequestBuilder struct {
	ctx     context.Context
	timeout time.Duration

	method       string
	baseURL      string
	pathElements []string
	header       http.Header
	query        urlpkg.Values
	body         struct {
		has bool
		v   any
	}
}

var _xReqBuilderPool = sync.Pool{
	New: func() any {
		return new(XRequestBuilder)
	},
}

func NewGet() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodGet
	return xr
}

func NewHead() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodHead
	return xr
}

func NewPost() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodPost
	return xr
}

func NewPut() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodPut
	return xr
}

func NewPatch() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodPatch
	return xr
}

func NewDelete() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodDelete
	return xr
}

func NewConnect() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodConnect
	return xr
}

func NewOptions() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodOptions
	return xr
}

func NewTrace() *XRequestBuilder {
	xr := _xReqBuilderPool.Get().(*XRequestBuilder)
	xr.method = http.MethodTrace
	return xr
}

func (xr *XRequestBuilder) WithContext(ctx context.Context) *XRequestBuilder {
	xr.ctx = ctx
	return xr
}

func (xr *XRequestBuilder) WithTimeout(d time.Duration) *XRequestBuilder {
	xr.timeout = d
	return xr
}

func (xr *XRequestBuilder) Header(header http.Header) *XRequestBuilder {
	xr.header = header.Clone()
	return xr
}

func (xr *XRequestBuilder) SetHeader(key, value string) *XRequestBuilder {
	if xr.header == nil {
		xr.header = make(http.Header)
	}
	xr.header.Set(key, value)
	return xr
}

func (xr *XRequestBuilder) AddHeader(key, value string) *XRequestBuilder {
	if xr.header == nil {
		xr.header = make(http.Header)
	}
	xr.header.Add(key, value)
	return xr
}

func (xr *XRequestBuilder) SetBasicAuth(username, password string) *XRequestBuilder {
	return xr.SetHeader("Authorization", "Basic "+basicAuth(username, password))
}

func (xr *XRequestBuilder) Query(query urlpkg.Values) *XRequestBuilder {
	xr.query = query
	return xr
}

func (xr *XRequestBuilder) SetQuery(key, value string) *XRequestBuilder {
	if xr.query == nil {
		xr.query = make(urlpkg.Values)
	}
	xr.query.Set(key, value)
	return xr
}

func (xr *XRequestBuilder) AddQuery(key, value string) *XRequestBuilder {
	if xr.query == nil {
		xr.query = make(urlpkg.Values)
	}
	xr.query.Add(key, value)
	return xr
}

func (xr *XRequestBuilder) Path(elements ...string) *XRequestBuilder {
	xr.pathElements = make([]string, 0, len(elements))
	for _, s := range elements {
		if s = strings.TrimSpace(s); s != "" {
			xr.pathElements = append(xr.pathElements, s)
		}
	}
	return xr
}

func (xr *XRequestBuilder) Body(body any) *XRequestBuilder {
	xr.body.has = true
	xr.body.v = body
	return xr
}

func (xr *XRequestBuilder) build(bc BodyCodec) (req *http.Request, cancel context.CancelFunc, err error) {
	defer xr.free()

	if xr.method == "" {
		panic("'XRequestBuilder' is not reusable")
	}

	cancel = func() {}

	u, err := xr.processingURL()
	if err != nil {
		return nil, cancel, fmt.Errorf("build url: %w", err)
	}

	br, err := xr.processingBody(bc)
	if err != nil {
		return nil, cancel, fmt.Errorf("build body: %w", err)
	}

	ctx := xr.ctx
	switch {
	case ctx == nil && xr.timeout <= 0:
		req, err = http.NewRequest(xr.method, u.String(), br)
	case ctx == nil && xr.timeout > 0:
		ctx, cancel = context.WithTimeout(context.Background(), xr.timeout)
		req, err = http.NewRequestWithContext(ctx, xr.method, u.String(), br)
	case ctx != nil && xr.timeout <= 0:
		req, err = http.NewRequestWithContext(ctx, xr.method, u.String(), br)
	case ctx != nil && xr.timeout > 0:
		ctx, cancel = context.WithTimeout(ctx, xr.timeout)
		req, err = http.NewRequestWithContext(ctx, xr.method, u.String(), br)
	}
	if err != nil {
		return nil, cancel, fmt.Errorf("build *http.Request: %w", err)
	}

	for k, v := range xr.header {
		req.Header[k] = append([]string{}, v...)
	}
	// req.Header = xr.header.Clone()

	return
}

func (xr *XRequestBuilder) processingURL() (u *urlpkg.URL, err error) {
	switch {
	case xr.baseURL == "" && len(xr.pathElements) == 0:
		return nil, errors.New("empty url")
	case xr.baseURL == "" && len(xr.pathElements) == 1:
		u, err = joinPath(xr.pathElements[0], nil)
	case xr.baseURL == "" && len(xr.pathElements) >= 2:
		u, err = joinPath(xr.pathElements[0], xr.pathElements[1:])
	case xr.baseURL != "" && len(xr.pathElements) == 0:
		u, err = joinPath(xr.baseURL, nil)
	default:
		// xr.baseURL != "" && len(xr.pathElements) >= 1
		var ref *urlpkg.URL
		if ref, err = urlpkg.Parse(xr.pathElements[0]); err != nil {
			return nil, err
		}
		if ref.IsAbs() {
			u = ref
			u, err = joinPath(u.String(), xr.pathElements[1:])
		} else {
			u, err = joinPath(xr.baseURL, xr.pathElements)
		}
	}
	if err != nil {
		return nil, err
	}

	if len(xr.query) != 0 {
		query, err := urlpkg.ParseQuery(u.RawQuery)
		if err != nil {
			return u, err
		}
		for k, vv := range xr.query {
			query[k] = vv
		}
		u.RawQuery = query.Encode()
	}

	return
}

func (xr *XRequestBuilder) processingBody(bc BodyCodec) (r io.Reader, err error) {
	if xr.body.has {
		if r, err = bc.Encode(xr.body.v); err != nil {
			return nil, err
		}
	}

	if bh, ok := bc.(BodyHeaderContentLength); ok {
		xr.SetHeader("Content-Length", strconv.Itoa(bh.ContentLength()))
	}
	if bh, ok := bc.(BodyHeaderContentType); ok {
		xr.SetHeader("Content-Type", bh.ContentType())
	}
	if bh, ok := bc.(BodyHeaderContentEncoding); ok {
		xr.SetHeader("Content-Encoding", bh.ContentEncoding())
	}

	if bh, ok := bc.(BodyHeaderAccept); ok {
		xr.SetHeader("Accept", bh.Accept())
	}
	if bh, ok := bc.(BodyHeaderAcceptEncoding); ok {
		xr.SetHeader("Accept-Encoding", bh.AcceptEncoding())
	}

	return
}

func (xr *XRequestBuilder) free() {
	xr.reset()
	_xReqBuilderPool.Put(xr)
}

func (xr *XRequestBuilder) reset() {
	xr.ctx = nil
	xr.timeout = 0
	xr.method = ""
	xr.baseURL = ""
	xr.pathElements = nil
	for k := range xr.header {
		xr.header.Del(k)
	}
	for k := range xr.query {
		xr.query.Del(k)
	}
	xr.body.has = false
	xr.body.v = nil
}
