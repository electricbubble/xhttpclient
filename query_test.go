package xhttpclient

import (
	"errors"
	"net/url"
	"testing"
)

func Test_xQuery_encode(t *testing.T) {
	tests := []struct {
		name     string
		fnNew    func() *xQuery
		argU     string
		argPatch string
		wantS    string
		wantErr  bool
	}{
		{
			name: "000",
			fnNew: func() *xQuery {
				return &xQuery{}
			},
			argU:     "",
			argPatch: "",
			wantS:    "",
			wantErr:  false,
		},
		{
			name: "000",
			fnNew: func() *xQuery {
				return &xQuery{}
			},
			argU:     "http://example",
			argPatch: "",
			wantS:    "",
			wantErr:  false,
		},
		{
			name: "100",
			fnNew: func() *xQuery {
				return &xQuery{}
			},
			argU:     "http://example?x=1",
			argPatch: "",
			wantS:    "x=1",
			wantErr:  false,
		},
		{
			name: "010",
			fnNew: func() *xQuery {
				return &xQuery{}
			},
			argU:     "http://example",
			argPatch: "?x=1",
			wantS:    "x=1",
			wantErr:  false,
		},
		{
			name: "110",
			fnNew: func() *xQuery {
				return &xQuery{}
			},
			argU:     "http://example?x=1",
			argPatch: "?y=2",
			wantS:    "x=1&y=2",
			wantErr:  false,
		},
		{
			name: "001",
			fnNew: func() *xQuery {
				q := &xQuery{}
				q.QueryAdd("x", "1")
				return q
			},
			argU:     "http://example",
			argPatch: "",
			wantS:    "x=1",
			wantErr:  false,
		},
		{
			name: "101",
			fnNew: func() *xQuery {
				q := &xQuery{}
				q.QueryAdd("x", "22")
				return q
			},
			argU:     "http://example?x=1",
			argPatch: "",
			wantS:    "x=1&x=22",
			wantErr:  false,
		},
		{
			name: "011",
			fnNew: func() *xQuery {
				q := &xQuery{}
				q.QueryAdd("x", "22")
				return q
			},
			argU:     "http://example",
			argPatch: "?x=1",
			wantS:    "x=1&x=22",
			wantErr:  false,
		},
		{
			name: "111",
			fnNew: func() *xQuery {
				q := &xQuery{}
				q.QueryAdd("y", "3")
				return q
			},
			argU:     "http://example?x=1",
			argPatch: "?y=2",
			wantS:    "x=1&y=2&y=3",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fnNew()
			u, err1 := url.Parse(tt.argU)
			patch, err2 := url.Parse(tt.argPatch)
			if err := errors.Join(err1, err2); err != nil {
				panic(err)
			}

			gotS, err := q.encode(u, patch)
			if (err != nil) != tt.wantErr {
				t.Errorf("encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotS != tt.wantS {
				t.Errorf("encode() gotS = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}
