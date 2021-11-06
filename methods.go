package xttp

import (
	"bytes"
	"github.com/hetiansu5/urlquery"
	mp "github.com/m-murad/ordered-sync-map"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var mu sync.RWMutex

// Get is a convenience helper for doing simple GET requests.
func (c *Client) Get(url string, body interface{}, headers *mp.Map) (*http.Response, error) {
	bts, _ := urlquery.Marshal(body)
	url = url + "?" + string(bts)
	return c.request(http.MethodGet, url, bytes.NewBufferString(""), headers)
}

// Get is a convenience helper for doing simple GET requests.
func (c *Client) GetJSON(url string, body interface{}, headers *mp.Map) (*http.Response, error) {
	headers.Put("Content-Type", "application/json")
	return c.Get(url, body, headers)
}

// Post is a convenience method for doing simple POST requests.
func (c *Client) Post(url string, body interface{}, headers *mp.Map) (*http.Response, error) {
	return c.request(http.MethodPost, url, body, headers)
}

// Post is a convenience method for doing simple POST requests.
func (c *Client) PostJSON(url string, body interface{}, headers *mp.Map) (*http.Response, error) {
	headers.Put("Content-Type", "application/json")
	return c.Post(url, body, headers)
}

// PostForm is a convenience method for doing simple POST operations using
// pre-filled url.Values form data.
func (c *Client) PostForm(url string, data url.Values, headers *mp.Map) (*http.Response, error) {
	headers.Put("Content-Type", "application/x-www-form-urlencoded")
	return c.Post(url, strings.NewReader(data.Encode()), headers)
}

// Get is a convenience helper for doing simple GET requests.
func (c *Client) Delete(url string, headers *mp.Map) (*http.Response, error) {
	return c.request(http.MethodDelete, url, nil, headers)
}

// Head is a convenience method for doing simple HEAD requests.
func (c *Client) Head(url string, headers *mp.Map) (*http.Response, error) {
	return c.request(http.MethodHead, url, nil, headers)
}

func (c *Client) request(method string, url string, body interface{}, headers *mp.Map) (*http.Response, error) {
	req, err := NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	headers.UnorderedRange(func(key interface{}, value interface{}) {
		req.Header.Set(key.(string), value.(string))
	})
	return c.Do(req)
}
