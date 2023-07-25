package xhttpclient

import (
	"bytes"
	"io"
	"net/http"
	urlpkg "net/url"
	"reflect"
	"strings"
	"testing"
)

// Edited from net/http/request_test.go

func TestXRequestBuilder_Build_URL_Query(t *testing.T) {
	type Want struct {
		urlNoQuery string
		query      urlpkg.Values
	}
	tests := []struct {
		name string
		xReq *XRequestBuilder
		base testTypeBase
		want Want
	}{
		{
			name: "has_base_1",
			xReq: NewGet().Path("search?hello=world"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/search",
				query:      urlpkg.Values{"hello": []string{"world"}},
			},
		},
		{
			name: "has_base_2",
			xReq: NewGet().Path("text", "search?hello=world"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/text/search",
				query:      urlpkg.Values{"hello": []string{"world"}},
			},
		},
		{
			name: "has_base_3",
			xReq: NewGet().
				Path("text", "search?q=foo&q=bar&hello=world").
				SetQuery("hello", "bye"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/text/search",
				query: urlpkg.Values{
					"q":     []string{"foo", "bar"},
					"hello": []string{"bye"},
				},
			},
		},
		{
			name: "has_base_4",
			xReq: NewGet().
				Path("text", "search?q=bar&hello=world").
				SetQuery("hello", "bye"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/text/search",
				query: urlpkg.Values{
					"q":     []string{"bar"},
					"hello": []string{"bye"},
				},
			},
		},
		{
			name: "has_base_5",
			xReq: NewGet().
				Path("text", "search?q=bar&hello=world").
				SetQuery("hello", "bye").
				SetQuery("q", "foo"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/text/search",
				query: urlpkg.Values{
					"q":     []string{"foo"},
					"hello": []string{"bye"},
				},
			},
		},
		{
			name: "has_base_6",
			xReq: NewGet().
				Path("text", "search?q=bar&hello=world").
				SetQuery("hello", "bye").
				SetQuery("q", "bar").
				AddQuery("q", "foo"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/text/search",
				query: urlpkg.Values{
					"q":     []string{"bar", "foo"},
					"hello": []string{"bye"},
				},
			},
		},
		{
			name: "has_base_7",
			xReq: NewGet().
				Path("test_data.json"),
			base: testTypeBase{
				url:    "https://www.example.com/search",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/search/test_data.json",
				query:      urlpkg.Values{},
			},
		},

		{
			name: "no_base_1",
			xReq: NewGet().Path("https://www.example.com/search?q=foo&q=bar"),
			base: testTypeBaseEmpty(),
			want: Want{
				urlNoQuery: "https://www.example.com/search",
				query:      urlpkg.Values{"q": []string{"foo", "bar"}},
			},
		},
		{
			name: "no_base_2",
			xReq: NewGet().Path("https://www.example.com", "search?q=foo&q=bar"),
			base: testTypeBaseEmpty(),
			want: Want{
				urlNoQuery: "https://www.example.com/search",
				query:      urlpkg.Values{"q": []string{"foo", "bar"}},
			},
		},
		{
			name: "no_base_3",
			xReq: NewGet().
				Path("https://www.example.com", "search?q=foo&q=bar").
				AddQuery("hello", "world"),
			base: testTypeBaseEmpty(),
			want: Want{
				urlNoQuery: "https://www.example.com/search",
				query: urlpkg.Values{
					"q":     []string{"foo", "bar"},
					"hello": []string{"world"},
				},
			},
		},
		{
			name: "no_base_4",
			xReq: NewGet().
				Path("https://www.example.com/search", "testdata_acc.json"),
			base: testTypeBaseEmpty(),
			want: Want{
				urlNoQuery: "https://www.example.com/search/testdata_acc.json",
				query:      urlpkg.Values{},
			},
		},

		{
			name: "both_1",
			xReq: NewGet().
				Path("https://www.example.com", "search?q=foo&q=bar").
				AddQuery("hello", "world"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/search",
				query: urlpkg.Values{
					"q":     []string{"foo", "bar"},
					"hello": []string{"world"},
				},
			},
		},

		{
			name: "both_2",
			xReq: NewGet().
				Path("https://www.example.com", "text", "search?q=foo&q=bar").
				AddQuery("hello", "world"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://www.example.com/text/search",
				query: urlpkg.Values{
					"q":     []string{"foo", "bar"},
					"hello": []string{"world"},
				},
			},
		},

		{
			name: "both_3",
			xReq: NewGet().
				Path("https://pkg.go.dev/search?q=github.com%2Felectricbubble&m=package"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://pkg.go.dev/search",
				query: urlpkg.Values{
					"q": []string{"github.com/electricbubble"},
					"m": []string{"package"},
				},
			},
		},

		{
			name: "both_4",
			xReq: NewGet().
				Path("https://pkg.go.dev/", "search?q=github.com%2Felectricbubble&m=package"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: Want{
				urlNoQuery: "https://pkg.go.dev/search",
				query: urlpkg.Values{
					"q": []string{"github.com/electricbubble"},
					"m": []string{"package"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTestXReq(tt.base, tt.xReq)
			bc := tt.base.codec.Get()
			defer tt.base.codec.Put(bc)
			req, err := tt.xReq.build(bc)
			if err != nil {
				t.Errorf("xReq.build() = %v; want nil", err)
				return
			}

			if s, _, _ := strings.Cut(req.URL.String(), "?"); s != tt.want.urlNoQuery {
				t.Errorf("req.URL(no query) = %s, want %v", s, tt.want.urlNoQuery)
				return
			}

			if !reflect.DeepEqual(req.URL.Query(), tt.want.query) {
				t.Errorf("req.URL.Query() = %v, wantQuery %v", req.URL.Query(), tt.want.query)
				return
			}
		})
	}
}

func TestXRequestBuilder_Build_Host(t *testing.T) {
	tests := []struct {
		name     string
		xReq     *XRequestBuilder
		base     testTypeBase
		wantHost string
	}{
		{
			name: "has_base_1",
			xReq: NewPost(),
			base: testTypeBase{
				url:    "http://www.example.com/",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			wantHost: "www.example.com",
		},
		{
			name: "has_base_2",
			xReq: NewPost(),
			base: testTypeBase{
				url:    "http://www.example.com:8080/",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			wantHost: "www.example.com:8080",
		},
		{
			name: "has_base_3",
			xReq: NewPost(),
			base: testTypeBase{
				url:    "http://192.168.0.1/",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			wantHost: "192.168.0.1",
		},
		{
			name: "has_base_4",
			xReq: NewPost(),
			base: testTypeBase{
				url:    "http://192.168.0.1:8080/",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			wantHost: "192.168.0.1:8080",
		},
		{
			name: "has_base_5",
			xReq: NewPost(),
			base: testTypeBase{
				url:    "http://192.168.0.1:/",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			wantHost: "192.168.0.1",
		},

		{
			name:     "no_base_1",
			xReq:     NewPost().Path("http://[fe80::1]/"),
			base:     testTypeBaseEmpty(),
			wantHost: "[fe80::1]",
		},
		{
			name:     "no_base_2",
			xReq:     NewPost().Path("http://[fe80::1]:8080/"),
			base:     testTypeBaseEmpty(),
			wantHost: "[fe80::1]:8080",
		},
		{
			name:     "no_base_3",
			xReq:     NewPost().Path("http://[fe80::1%25en0]/"),
			base:     testTypeBaseEmpty(),
			wantHost: "[fe80::1%en0]",
		},
		{
			name:     "no_base_4",
			xReq:     NewPost().Path("http://[fe80::1%25en0]:8080/"),
			base:     testTypeBaseEmpty(),
			wantHost: "[fe80::1%en0]:8080",
		},
		{
			name:     "no_base_5",
			xReq:     NewPost().Path("http://[fe80::1%25en0]:/"),
			base:     testTypeBaseEmpty(),
			wantHost: "[fe80::1%en0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTestXReq(tt.base, tt.xReq)
			bc := tt.base.codec.Get()
			defer tt.base.codec.Put(bc)
			req, err := tt.xReq.build(bc)
			if err != nil {
				t.Errorf("xReq.build() = %v; want nil", err)
				return
			}

			if req.Host != tt.wantHost {
				t.Errorf("req.Host = %v, want %v", req.Host, tt.wantHost)
				return
			}
		})
	}
}

type testTypeBasicAuth struct {
	username, password string
	ok                 bool
}

func TestXRequestBuilder_Build_BasicAuth(t *testing.T) {
	tests := []struct {
		name string
		xReq *XRequestBuilder
		base testTypeBase
		want testTypeBasicAuth
	}{
		{
			name: "has_base_1",
			xReq: NewPut().SetBasicAuth("golang", "hello"),
			base: testTypeBase{
				url: "https://www.example.com",
				header: func() http.Header {
					h := make(http.Header)
					h.Set("Authorization", "Basic "+basicAuth("golang", "hello"))
					return h
				}(),
				codec: testFakeBodyCodecJSONEmpty(),
			},
			want: testTypeBasicAuth{
				username: "golang",
				password: "hello",
				ok:       true,
			},
		},
		{
			name: "has_base_2",
			xReq: NewPut().SetBasicAuth("", ""),
			base: testTypeBase{
				url: "https://www.example.com",
				header: func() http.Header {
					h := make(http.Header)
					h.Set("Authorization", "Basic "+basicAuth("", ""))
					return h
				}(),
				codec: testFakeBodyCodecJSONEmpty(),
			},
			want: testTypeBasicAuth{
				username: "",
				password: "",
				ok:       true,
			},
		},

		{
			name: "no_base_1",
			xReq: NewPut().SetBasicAuth("x", "go"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: testTypeBasicAuth{
				username: "x",
				password: "go",
				ok:       true,
			},
		},
		{
			name: "no_base_2",
			xReq: NewHead(),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  testFakeBodyCodecJSONEmpty(),
			},
			want: testTypeBasicAuth{
				username: "",
				password: "",
				ok:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTestXReq(tt.base, tt.xReq)
			bc := tt.base.codec.Get()
			defer tt.base.codec.Put(bc)
			req, err := tt.xReq.build(bc)
			if err != nil {
				t.Errorf("xReq.build() = %v; want nil", err)
				return
			}

			username, password, ok := req.BasicAuth()
			if ok != tt.want.ok || username != tt.want.username || password != tt.want.password {
				t.Errorf("BasicAuth() = %+v, want %+v", testTypeBasicAuth{username, password, ok},
					testTypeBasicAuth{tt.want.username, tt.want.password, tt.want.ok})
				return
			}
		})
	}

}

var _ BodyCodec = (*testFakeBodyCodecJSON)(nil)

type testFakeBodyCodecJSON struct {
	fnEncode func(v any) (io.Reader, error)
	fnDecode func(r io.Reader, v any) error
}

func testFakeBodyCodecJSONEmpty() *testFakeBodyCodecJSON {
	return &testFakeBodyCodecJSON{
		fnEncode: func(v any) (io.Reader, error) { return nil, nil },
		fnDecode: func(r io.Reader, v any) error { return nil },
	}
}

func (t *testFakeBodyCodecJSON) Get() BodyCodec {
	return t
}

func (t *testFakeBodyCodecJSON) Put(bc BodyCodec) {}

func (t *testFakeBodyCodecJSON) Encode(body any) (io.Reader, error) {
	return t.fnEncode(body)
}

func (t *testFakeBodyCodecJSON) Decode(r io.Reader, v any) error {
	return t.fnDecode(r, v)
}

func TestXRequestBuilder_Build_GetBody(t *testing.T) {
	tests := []struct {
		name     string
		xReq     *XRequestBuilder
		base     testTypeBase
		wantBody string
	}{
		{
			name: "1",
			xReq: NewPatch().Body("hello"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec: &testFakeBodyCodecJSON{
					fnEncode: func(v any) (io.Reader, error) {
						return strings.NewReader(v.(string)), nil
					},
				},
			},
			wantBody: "hello",
		},
		{
			name: "2",
			xReq: NewPatch().Body("golang"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec: &testFakeBodyCodecJSON{
					fnEncode: func(v any) (io.Reader, error) {
						return bytes.NewReader([]byte(v.(string))), nil
					},
				},
			},
			wantBody: "golang",
		},
		{
			name: "3",
			xReq: NewPatch().Body("x"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec: &testFakeBodyCodecJSON{
					fnEncode: func(v any) (io.Reader, error) {
						return bytes.NewBuffer([]byte(v.(string))), nil
					},
				},
			},
			wantBody: "x",
		},
		{
			name: "4",
			xReq: NewPatch().Body("bye"),
			base: testTypeBase{
				url:    "https://www.example.com",
				header: nil,
				codec:  BodyCodecJSON,
			},
			wantBody: `"bye"` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTestXReq(tt.base, tt.xReq)
			bc := tt.base.codec.Get()
			defer tt.base.codec.Put(bc)
			req, err := tt.xReq.build(bc)
			if err != nil {
				t.Errorf("xReq.build() = %v; want nil", err)
				return
			}

			if req.Body == nil {
				t.Error("Body = nil")
				return
			}
			if req.GetBody == nil {
				t.Error("GetBody = nil")
				return
			}
			slurp1, err := io.ReadAll(req.Body)
			if err != nil {
				t.Errorf("ReadAll(Body) = %v", err)
				return
			}
			newBody, err := req.GetBody()
			if err != nil {
				t.Errorf("GetBody = %v", err)
				return
			}
			slurp2, err := io.ReadAll(newBody)
			if err != nil {
				t.Errorf("ReadAll(GetBody()) = %v", err)
				return
			}
			if string(slurp1) != string(slurp2) {
				t.Errorf("Body %q != GetBody %q", slurp1, slurp2)
				return
			}
			if string(slurp1) != tt.wantBody {
				t.Errorf("Body %q != want %q", slurp1, tt.wantBody)
				return
			}
		})
	}
}
