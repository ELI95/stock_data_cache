package utils

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type RequestOption func(*http.Request)

func RequestWithHeaders(headers map[string]string) RequestOption {
	return func(req *http.Request) {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
}

func switchContentEncoding(resp *http.Response) (bodyReader io.Reader, err error) {
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		bodyReader, err = gzip.NewReader(resp.Body)
	case "deflate":
		bodyReader = flate.NewReader(resp.Body)
	default:
		bodyReader = resp.Body
	}
	return
}

func DoGetRequest(url string, timeout time.Duration, opts ...RequestOption) (b []byte, err error) {
	defer TimeTrack(time.Now(), "DoGetRequest")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	for _, opt := range opts {
		opt(req)
	}

	// fix EOF error
	// it prevents the connection from being re-used
	// see https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi/23963271
	req.Close = true

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("status code: %d", resp.StatusCode))
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	bodyReader, err := switchContentEncoding(resp)
	b, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		return
	}
	return
}

func DoPostRequest(url string, timeout time.Duration, body io.Reader, opts ...RequestOption,) (b []byte, err error) {
	defer TimeTrack(time.Now(), "DoPostRequest")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return
	}

	for _, opt := range opts {
		opt(req)
	}

	// fix EOF error
	// it prevents the connection from being re-used
	// see https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi/23963271
	req.Close = true

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("status code: %d", resp.StatusCode))
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	bodyReader, err := switchContentEncoding(resp)
	b, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		return
	}
	return
}
