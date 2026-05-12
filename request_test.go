package xhttpclient

import (
	"reflect"
	"testing"
)

func Test_xRequestBuilder_buildURL(t *testing.T) {
	tests := []struct {
		name    string
		fnNew   func() RequestBuilder
		wantU   string
		wantErr bool
	}{
		{
			name: "",
			fnNew: func() RequestBuilder {
				return NewGet()
			},
			wantU:   "",
			wantErr: true,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				return NewGet().
					Path("http://example")
			},
			wantU:   "http://example",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				return NewGet().
					Path("http://example", "p1")
			},
			wantU:   "http://example/p1",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				return NewGet().
					Path("http://example/p1/")
			},
			wantU:   "http://example/p1/",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				r := NewGet().(*xRequestBuilder)
				r.baseURL = "http://example"
				return r
			},
			wantU:   "http://example",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				r := NewGet().(*xRequestBuilder)
				r.baseURL = "http://example"
				return r.Path("https://example.com")
			},
			wantU:   "https://example.com",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				r := NewGet().(*xRequestBuilder)
				r.baseURL = "http://example"
				return r.Path("https://example.com", "p1%2Fs2")
			},
			wantU:   "https://example.com/p1%2Fs2",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				r := NewGet().(*xRequestBuilder)
				r.baseURL = "http://example?x=1"
				return r.Path("http://example", "p1").
					QueryAdd("y", "2")
			},
			wantU:   "http://example/p1?y=2",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				r := NewGet().(*xRequestBuilder)
				r.baseURL = "http://example"
				return r.Path("p1", "p2?x=1")
			},
			wantU:   "http://example/p1/p2?x=1",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				r := NewGet().(*xRequestBuilder)
				r.baseURL = "http://example"
				return r.Path("p1", "p2%2Fs3?x=1")
			},
			wantU:   "http://example/p1/p2%2Fs3?x=1",
			wantErr: false,
		},
		{
			name: "",
			fnNew: func() RequestBuilder {
				r := NewGet().(*xRequestBuilder)
				r.baseURL = "http://example"
				return r.Path("http://example/p1/")
			},
			wantU:   "http://example/p1/",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fnNew().(*xRequestBuilder)
			gotU, err := r.buildURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("buildURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				t.Logf("buildURL(), expected error: %v", err)
				return
			}
			if !reflect.DeepEqual(gotU.String(), tt.wantU) {
				t.Errorf("buildURL() gotU = %v, want %v", gotU, tt.wantU)
			}
		})
	}
}
