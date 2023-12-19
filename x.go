package xhttpclient

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"path"
	"sync"
)

var _bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func getBytesBuffer() *bytes.Buffer {
	return _bufPool.Get().(*bytes.Buffer)
}

func putBytesBuffer(buf *bytes.Buffer) {
	buf.Reset()
	_bufPool.Put(buf)
}

func joinPath(baseURL string, pathElements []string) (u *urlpkg.URL, err error) {
	if u, err = urlpkg.Parse(baseURL); err != nil {
		return nil, err
	}
	if len(pathElements) == 0 {
		return u, nil
	}

	pu, err := urlpkg.Parse(path.Join(pathElements...))
	if err != nil {
		return nil, err
	}

	u = u.JoinPath(pu.EscapedPath())
	u.RawQuery = pu.RawQuery
	return u, nil
}

// See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
//
// Copied from net/http/client.go
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Copied from mime/multipart/writer.go
func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func wrapCancelAndCloseRespBody(cancel context.CancelFunc, resp *http.Response) context.CancelFunc {
	if resp == nil {
		return cancel
	}
	return func() {
		cancel()
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}
