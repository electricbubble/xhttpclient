package xhttpclient

import (
	"encoding/base64"
	"net/http"
)

type xHeader struct {
	header http.Header
}

func (h *xHeader) Header(header http.Header) {
	h.header = header
}

func (h *xHeader) HeaderSet(key, value string) {
	if h.header == nil {
		h.header = make(http.Header)
	}
	h.header.Set(key, value)
}

func (h *xHeader) HeaderAdd(key, value string) {
	if h.header == nil {
		h.header = make(http.Header)
	}
	h.header.Add(key, value)
}

func (h *xHeader) HeaderSetBasicAuth(username, password string) {
	// copied from net/http/client.go
	auth := username + ":" + password
	s := base64.StdEncoding.EncodeToString([]byte(auth))

	h.HeaderSet("Authorization", "Basic "+s)
}

func (h *xHeader) apply(patch xHeader) {
	switch n := len(patch.header); {
	case n == 0:
		return

	case h.header == nil:
		h.header = make(http.Header, n)
	}

	for k, vv := range patch.header {
		for _, v := range vv {
			h.header.Add(k, v)
		}
	}
}
