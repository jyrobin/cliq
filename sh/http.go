package sh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Response holds HTTP response data.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       string
	Err        error
}

// OK returns true if status is 2xx.
func (r *Response) OK() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// JSON parses the body as JSON into a generic map.
func (r *Response) JSON() (map[string]interface{}, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(r.Body), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// JSONArray parses the body as JSON array.
func (r *Response) JSONArray() ([]interface{}, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	var arr []interface{}
	if err := json.Unmarshal([]byte(r.Body), &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// HTTPClient wraps http.Client with convenient methods.
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// HTTP creates a new HTTP client.
func HTTP() *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		headers: make(map[string]string),
	}
}

// BaseURL sets the base URL for all requests.
func (c *HTTPClient) BaseURL(url string) *HTTPClient {
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

// Header sets a default header for all requests.
func (c *HTTPClient) Header(key, value string) *HTTPClient {
	c.headers[key] = value
	return c
}

// Timeout sets the client timeout.
func (c *HTTPClient) Timeout(d time.Duration) *HTTPClient {
	c.client.Timeout = d
	return c
}

// Auth sets Bearer token authentication.
func (c *HTTPClient) Auth(token string) *HTTPClient {
	c.headers["Authorization"] = "Bearer " + token
	return c
}

// BasicAuth sets Basic authentication.
func (c *HTTPClient) BasicAuth(user, pass string) *HTTPClient {
	// Will be handled specially in request
	c.headers["_basic_auth"] = user + ":" + pass
	return c
}

func (c *HTTPClient) buildURL(path string) string {
	if c.baseURL == "" {
		return path
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return c.baseURL + "/" + strings.TrimLeft(path, "/")
}

func (c *HTTPClient) do(req *http.Request) *Response {
	// Apply default headers
	for k, v := range c.headers {
		if k == "_basic_auth" {
			parts := strings.SplitN(v, ":", 2)
			if len(parts) == 2 {
				req.SetBasicAuth(parts[0], parts[1])
			}
			continue
		}
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return &Response{Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{Err: err}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       string(body),
	}
}

// Get performs a GET request.
func (c *HTTPClient) Get(path string) *Response {
	return c.GetContext(context.Background(), path)
}

// GetContext performs a GET request with context.
func (c *HTTPClient) GetContext(ctx context.Context, path string) *Response {
	req, err := http.NewRequestWithContext(ctx, "GET", c.buildURL(path), nil)
	if err != nil {
		return &Response{Err: err}
	}
	return c.do(req)
}

// Post performs a POST request with body.
func (c *HTTPClient) Post(path string, body string) *Response {
	return c.PostContext(context.Background(), path, body)
}

// PostContext performs a POST request with context.
func (c *HTTPClient) PostContext(ctx context.Context, path string, body string) *Response {
	req, err := http.NewRequestWithContext(ctx, "POST", c.buildURL(path), strings.NewReader(body))
	if err != nil {
		return &Response{Err: err}
	}
	return c.do(req)
}

// PostJSON performs a POST request with JSON body.
func (c *HTTPClient) PostJSON(path string, data interface{}) *Response {
	return c.PostJSONContext(context.Background(), path, data)
}

// PostJSONContext performs a POST request with JSON body and context.
func (c *HTTPClient) PostJSONContext(ctx context.Context, path string, data interface{}) *Response {
	body, err := json.Marshal(data)
	if err != nil {
		return &Response{Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.buildURL(path), bytes.NewReader(body))
	if err != nil {
		return &Response{Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// PostForm performs a POST request with form data.
func (c *HTTPClient) PostForm(path string, data map[string]string) *Response {
	return c.PostFormContext(context.Background(), path, data)
}

// PostFormContext performs a POST request with form data and context.
func (c *HTTPClient) PostFormContext(ctx context.Context, path string, data map[string]string) *Response {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.buildURL(path), strings.NewReader(form.Encode()))
	if err != nil {
		return &Response{Err: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req)
}

// Put performs a PUT request.
func (c *HTTPClient) Put(path string, body string) *Response {
	return c.PutContext(context.Background(), path, body)
}

// PutContext performs a PUT request with context.
func (c *HTTPClient) PutContext(ctx context.Context, path string, body string) *Response {
	req, err := http.NewRequestWithContext(ctx, "PUT", c.buildURL(path), strings.NewReader(body))
	if err != nil {
		return &Response{Err: err}
	}
	return c.do(req)
}

// PutJSON performs a PUT request with JSON body.
func (c *HTTPClient) PutJSON(path string, data interface{}) *Response {
	return c.PutJSONContext(context.Background(), path, data)
}

// PutJSONContext performs a PUT request with JSON body and context.
func (c *HTTPClient) PutJSONContext(ctx context.Context, path string, data interface{}) *Response {
	body, err := json.Marshal(data)
	if err != nil {
		return &Response{Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", c.buildURL(path), bytes.NewReader(body))
	if err != nil {
		return &Response{Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// Delete performs a DELETE request.
func (c *HTTPClient) Delete(path string) *Response {
	return c.DeleteContext(context.Background(), path)
}

// DeleteContext performs a DELETE request with context.
func (c *HTTPClient) DeleteContext(ctx context.Context, path string) *Response {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.buildURL(path), nil)
	if err != nil {
		return &Response{Err: err}
	}
	return c.do(req)
}

// Request performs a custom request.
func (c *HTTPClient) Request(method, path string, body io.Reader) *Response {
	return c.RequestContext(context.Background(), method, path, body)
}

// RequestContext performs a custom request with context.
func (c *HTTPClient) RequestContext(ctx context.Context, method, path string, body io.Reader) *Response {
	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path), body)
	if err != nil {
		return &Response{Err: err}
	}
	return c.do(req)
}

// Convenience functions for quick one-off requests

// Get performs a simple GET request.
func Get(url string) *Response {
	return HTTP().Get(url)
}

// Post performs a simple POST request.
func Post(url string, body string) *Response {
	return HTTP().Post(url, body)
}

// PostJSON performs a simple POST request with JSON body.
func PostJSON(url string, data interface{}) *Response {
	return HTTP().PostJSON(url, data)
}

// GetJSON performs GET and parses response as JSON.
func GetJSON(url string) (map[string]interface{}, error) {
	resp := Get(url)
	if resp.Err != nil {
		return nil, resp.Err
	}
	if !resp.OK() {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return resp.JSON()
}
