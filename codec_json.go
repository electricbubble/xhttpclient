package xhttpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type xJSON struct {
	EncoderJSON
	DecoderJSON
}

type EncoderJSON struct{}

func (c *EncoderJSON) Encode(m RequestMutation) (body io.Reader, err error) {
	v, ok := m.GetBody()
	if !ok {
		return nil, nil
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err = enc.Encode(v); err != nil {
		return nil, err
	}

	m.HeaderSet(HeaderKeyContentType, HeaderValueJSON)
	m.HeaderSet(HeaderKeyContentLength, strconv.Itoa(buf.Len()))

	body = buf
	return
}

type DecoderJSON struct{}

func (c *DecoderJSON) Decode(resp ResponseHandle, successV, failureV any) (err error) {
	respStatusCode := resp.StatusCode()
	if err = resp.BodyReadAll(); err != nil {
		return fmt.Errorf("body: read: %w", err)
	}
	respBody := resp.BodyHasBeenRead()

	switch {
	case respStatusCode == http.StatusNoContent:
		return nil

	case 200 <= respStatusCode && respStatusCode <= 299:
		if err = json.Unmarshal(respBody, successV); err != nil {
			err = fmt.Errorf("json: unmarshal: %w", err)
		}

	case failureV != nil:
		if err = json.Unmarshal(respBody, failureV); err != nil {
			err = fmt.Errorf("json: unmarshal: %w", err)
		}

	default:
		err = ErrUnhandled
	}

	if err != nil {
		err = fmt.Errorf("%w\n"+
			"Method: %s\nURL: %s\n"+
			"StatusCode: %d\n"+
			"Body(Len: %d): %s",
			err,
			resp.RequestMethod(), resp.RequestURL(),
			respStatusCode,
			len(respBody), respBody,
		)
	}
	return err
}
