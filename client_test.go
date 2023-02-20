package xhttpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkXClient_Do(b *testing.B) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()
	cli := NewClient()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := cli.Do(nil, nil, NewGet().Path(ts.URL))
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func TestXClient_Do_FormUrlencoded(t *testing.T) {
	type Response struct {
		Form struct {
			Name  string `json:"name"`
			Tel   string `json:"tel"`
			Email string `json:"email"`
		} `json:"form"`
		Headers struct {
			Accept        string `json:"Accept"`
			ContentLength string `json:"Content-Length"`
			ContentType   string `json:"Content-Type"`
		} `json:"headers"`
	}

	cli := NewClient().BaseURL("https://httpbin.org")

	name, tel, email := "hi", "123", "456@789.com"
	formData := urlpkg.Values{}
	formData.Set("name", name)
	formData.Set("tel", tel)
	formData.Set("email", email)

	var (
		successV Response
		wrongV   map[string]any
	)
	resp, respBody, err := cli.DoOnceWithBodyCodec(BodyCodecFormUrlencodedAndJSON, &successV, &wrongV,
		NewPost().
			Path("/post").
			Body(formData),
	)
	if err != nil {
		t.Fatalf("%s\n%s", err, respBody)
	}
	if len(wrongV) != 0 {
		t.Fatal(fmt.Errorf("wrong response (HTTP status: %s(%d)): %+v", http.StatusText(resp.StatusCode), resp.StatusCode, wrongV))
	}

	switch {
	case successV.Form.Name != name:
		t.Fatalf("form.name = %s, want %v", successV.Form.Name, name)
	case successV.Form.Tel != tel:
		t.Fatalf("form.tel = %s, want %v", successV.Form.Tel, tel)
	case successV.Form.Email != email:
		t.Fatalf("form.email = %s, want %v", successV.Form.Email, email)
	case successV.Headers.ContentType != ContentTypeValueFormUrlencoded:
		t.Fatalf("header.ContentType = %s, want %v", successV.Headers.ContentType, ContentTypeValueFormUrlencoded)
	case successV.Headers.Accept != ContentTypeValueJSON:
		t.Fatalf("header.Accept = %s, want %v", successV.Headers.Accept, ContentTypeValueJSON)
	case successV.Headers.ContentLength != strconv.Itoa(len([]byte(formData.Encode()))):
		t.Fatalf("header.ContentLength = %s, want %v", successV.Headers.ContentLength, strconv.Itoa(len([]byte(formData.Encode()))))
	}

	bs, err := json.MarshalIndent(successV, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("successV\n", string(bs))

}

func TestXClient_Do_Multipart(t *testing.T) {
	type Response struct {
		Files struct {
			F1 string `json:"f1"`
		} `json:"files"`
		Form struct {
			K1 string `json:"k1"`
			K2 string `json:"k2"`
		} `json:"form"`
		Headers struct {
			Accept        string `json:"Accept"`
			ContentLength string `json:"Content-Length"`
			ContentType   string `json:"Content-Type"`
		} `json:"headers"`
	}

	cli := NewClient().BaseURL("https://httpbin.org")

	v1, v2, fContent := "world", "hi again", `{"x","go"}`

	tmpTestdata := filepath.Join(t.TempDir(), "testdata.json")
	if err := os.WriteFile(tmpTestdata, []byte(fContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpTestdata)

	mw := NewMultipartWriter().
		WriteWithFieldValue("k1", v1).
		WriteWithField("k2", strings.NewReader(v2)).
		WriteWithFile("f1", tmpTestdata)
	contentType := mw.FormDataContentType()

	var (
		successV Response
		wrongV   map[string]any
	)
	resp, respBody, err := cli.DoOnceWithBodyCodec(BodyCodecMultipart, &successV, &wrongV,
		NewPost().
			Path("/anything").
			Body(mw),
	)
	if err != nil {
		t.Fatalf("%s\n%s", err, respBody)
	}
	if len(wrongV) != 0 {
		t.Fatal(fmt.Errorf("wrong response (HTTP status: %s(%d)): %+v", http.StatusText(resp.StatusCode), resp.StatusCode, wrongV))
	}

	switch {
	case successV.Form.K1 != v1:
		t.Fatalf("form.k1 = %s, want %v", successV.Form.K1, v1)
	case successV.Form.K2 != v2:
		t.Fatalf("form.k2 = %s, want %v", successV.Form.K2, v2)
	case successV.Files.F1 != fContent:
		t.Fatalf("files.f1 = %s, want %v", successV.Files.F1, fContent)
	case successV.Headers.ContentType != contentType:
		t.Fatalf("header.ContentType = %s, want %v", successV.Headers.ContentType, contentType)
	case successV.Headers.Accept != ContentTypeValueJSON:
		t.Fatalf("header.Accept = %s, want %v", successV.Headers.Accept, ContentTypeValueJSON)
	}

	buf := new(bytes.Buffer)
	err = NewMultipartWriter().
		WriteWithFieldValue("k1", v1).
		WriteWithField("k2", strings.NewReader(v2)).
		WriteWithFile("f1", tmpTestdata).
		do(buf)
	if err != nil {
		t.Fatal(err)
	}
	if successV.Headers.ContentLength != strconv.Itoa(buf.Len()) {
		t.Fatalf("header.ContentLength = %s, want %v", successV.Headers.ContentLength, strconv.Itoa(buf.Len()))
	}

	bs, err := json.MarshalIndent(successV, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("successV\n", string(bs))
}

func TestXClient_Do_gzip(t *testing.T) {
	cli := NewClient().BaseURL("https://httpbin.org")

	var (
		successV any
		wrongV   map[string]any
	)
	resp, respBody, err := cli.Do(&successV, &wrongV, NewGet().Path("/gzip"))
	if err != nil {
		t.Fatalf("%s\n%s", err, respBody)
	}
	if len(wrongV) != 0 {
		t.Fatal(fmt.Errorf("wrong response (HTTP status: %s(%d)): %+v", http.StatusText(resp.StatusCode), resp.StatusCode, wrongV))
	}

	bs, err := json.MarshalIndent(successV, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("successV\n", string(bs))
}
