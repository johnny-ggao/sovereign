package cobo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type httpClient struct {
	baseURL    string
	signer     *signer
	httpClient *http.Client
}

func newHTTPClient(baseURL string, s *signer) *httpClient {
	return &httpClient{
		baseURL: baseURL,
		signer:  s,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *httpClient) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	queryStr := ""
	fullPath := path
	if len(params) > 0 {
		queryStr = params.Encode()
		fullPath = path + "?" + queryStr
	}

	return c.do(ctx, http.MethodGet, path, queryStr, fullPath, nil)
}

func (c *httpClient) post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	return c.do(ctx, http.MethodPost, path, "", path, payload)
}

// do 执行 HTTP 请求，带重试。
// signPath 用于签名（不含 query string），fullPath 用于实际请求 URL。
func (c *httpClient) do(ctx context.Context, method, signPath, params, fullPath string, body []byte) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, err := c.doOnce(ctx, method, signPath, params, fullPath, body)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if !isRetryable(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *httpClient) doOnce(ctx context.Context, method, signPath, params, fullPath string, body []byte) ([]byte, error) {
	reqURL := c.baseURL + fullPath

	var bodyReader io.Reader
	bodyStr := ""
	if body != nil {
		bodyReader = bytes.NewReader(body)
		bodyStr = string(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	nonce, sig := c.signer.sign(method, signPath, params, bodyStr)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Biz-Api-Key", c.signer.pubHex)
	req.Header.Set("Biz-Api-Nonce", nonce)
	req.Header.Set("Biz-Api-Signature", sig)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.ErrorCode != 0 {
			apiErr.HTTPStatus = resp.StatusCode
			return nil, &apiErr
		}
		return nil, &APIError{
			HTTPStatus: resp.StatusCode,
			ErrorCode:  resp.StatusCode,
			ErrorMsg:   string(respBody),
		}
	}

	return respBody, nil
}
