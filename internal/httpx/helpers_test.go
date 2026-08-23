package httpx

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
)

func newGet(target string) *http.Request {
	return httptest.NewRequest(http.MethodGet, target, nil)
}

func newPost(target string) *http.Request {
	return httptest.NewRequest(http.MethodPost, target, nil)
}

func newDelete(target string) *http.Request {
	return httptest.NewRequest(http.MethodDelete, target, nil)
}

func bodyBytes(resp *http.Response) []byte {
	if resp.Body == nil {
		return nil
	}
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)
	return buf.Bytes()
}
