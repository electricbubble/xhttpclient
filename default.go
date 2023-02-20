package xhttpclient

import (
	"net"
	"net/http"
	"runtime"
	"time"
)

func DefaultClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport(),
		Timeout:   30 * time.Second,
	}
}

func DefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   runtime.NumCPU() + 1,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
