package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GetJSON performs a GET request with retries and returns the parsed JSON response.
func GetJSON(rawURL string, headers map[string]string, params map[string]string) (map[string]any, error) {
	return GetJSONWithOptions(rawURL, headers, params, 3, 500*time.Millisecond, 60*time.Second)
}

// GetJSONWithOptions performs a GET request with configurable retry/timeout settings.
func GetJSONWithOptions(
	rawURL string,
	headers map[string]string,
	params map[string]string,
	retries int,
	retrySleep time.Duration,
	timeout time.Duration,
) (map[string]any, error) {
	if len(params) > 0 {
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL %q: %w", rawURL, err)
		}
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		rawURL = u.String()
	}

	client := &http.Client{Timeout: timeout}

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		req, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < retries {
				time.Sleep(retrySleep)
				continue
			}
			return nil, fmt.Errorf("HTTP request failed after %d retries: %w", retries, lastErr)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			if attempt < retries {
				time.Sleep(retrySleep)
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
			if attempt < retries && resp.StatusCode >= 500 {
				time.Sleep(retrySleep)
				continue
			}
			return nil, lastErr
		}

		var result map[string]any
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing JSON response: %w", err)
		}
		return result, nil
	}
	return nil, lastErr
}

// HTTPError represents an HTTP response with a non-2xx status code.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}
