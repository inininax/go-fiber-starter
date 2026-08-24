package health

import (
	"io"
	"net/http"
	"net/http/httptest"
)

func request(method, target string) *http.Request {
	return httptest.NewRequest(method, target, nil)
}

func readAll(resp *http.Response) []byte {
	if resp.Body == nil {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return body
}
