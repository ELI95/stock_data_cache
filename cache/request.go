package cache

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/axgle/mahonia"
	"io"
	"io/ioutil"
	"net/http"
	"stock_data_cache/utils"
	"time"
)

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

func RequestSina(url string) (data string, err error) {
	defer utils.TimeTrack(time.Now(), "RequestSina")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// fix EOF error
	// it prevents the connection from being re-used
	// see https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi/23963271
	req.Close = true
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://finance.sina.com.cn/")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req.WithContext(ctx)

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
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return
	}
	data = mahonia.NewDecoder("gbk").ConvertString(string(body))
	return
}

func DoGetRequest(url string) (data string, err error) {
	defer utils.TimeTrack(time.Now(), "DoGetRequest")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// fix EOF error
	// it prevents the connection from being re-used
	// see https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi/23963271
	req.Close = true
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req.WithContext(ctx)

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	data = string(body)
	return
}

func DoPostRequest(url string, body io.Reader) (*http.Response, error) {
	defer utils.TimeTrack(time.Now(), "DoPostRequest")

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	// fix EOF error
	// it prevents the connection from being re-used
	// see https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi/23963271
	req.Close = true
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}