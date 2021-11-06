package xttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	mp "github.com/m-murad/ordered-sync-map"
	"github.com/phuslu/log"
	"github.com/rs/xid"
	"github.com/sujit-baniya/xttp/pool"
)

type HTTPRequest struct {
	ID      string
	URL     string
	Payload interface{}
	Headers      *mp.Map
	RetryMax     int
	RetryWaitMax time.Duration
	Timeout      time.Duration
	ReqPerSec    int
	MaxPoolSize  int
	Response     *http.Response
	HttpError    error
	Status       int
	client       *Client
	LogRequest   bool
}

func (w *HTTPRequest) Client() *Client {
	if w.client == nil {
		opts := Options{
			RetryWaitMax: w.RetryWaitMax,
			Timeout:      w.Timeout,
			RetryMax:     w.RetryMax,
			MaxPoolSize:  w.MaxPoolSize,
			KillIdleConn: true,
			ReqPerSec:    w.ReqPerSec,
		}
		w.client = NewClient(opts)
	}
	return w.client
}

func (w *HTTPRequest) Get(payload interface{}) *HTTPRequest {
	var y *HTTPRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().Get(w.URL, payload, w.Headers)
	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}

	y.Log(payload, nil)
	return y
}

func (w *HTTPRequest) Log(payload interface{}, err error) {
	if !w.LogRequest {
		return
	}
	req, _ := json.Marshal(payload)
	mu.RLock()

	var e *log.Entry
	response := ""
	statusCode := 20
	if w.Response != nil {
		resp, _ := ioutil.ReadAll(w.Response.Body)
		response = string(resp)
		statusCode = w.Response.StatusCode
		e = log.Info()
	} else {
		e = log.Error()
	}

	if err != nil {
		e = log.Error()
	} else {
		e = log.Info()
	}
	headers := make(map[string]string)
	w.Headers.UnorderedRange(func(key interface{}, value interface{}) {
		headers[key.(string)] = value.(string)
	})
	head, _ := json.Marshal(headers)
	e.
		Str("request_id", xid.New().String()).
		Str("url", w.URL).
		RawJSON("request_payload", req).
		RawJSON("request_header", head).
		Int("status", statusCode).
		Str("response", response).
		Msg("Client Response")
	mu.RUnlock()
}

func (w *HTTPRequest) PostJson(payload []byte) *HTTPRequest {
	var y *HTTPRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().PostJSON(w.URL, bytes.NewBufferString(string(payload)), w.Headers)
	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}
	y.Log(payload, nil)
	return y
}

func (w *HTTPRequest) GetJson(payload interface{}) *HTTPRequest {
	var y *HTTPRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().GetJSON(w.URL, payload, w.Headers)
	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}

	y.Log(payload, nil)
	return y
}

func (w *HTTPRequest) Post(payload interface{}) *HTTPRequest {
	var y *HTTPRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().Post(w.URL, payload, w.Headers)

	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}

	y.Log(payload, nil)
	return y
}

func (w *HTTPRequest) PostForm(payload url.Values) *HTTPRequest {
	var y *HTTPRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().PostForm(w.URL, payload, w.Headers)

	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}

	y.Log(payload, nil)
	return y
}

func (w *HTTPRequest) AsyncGet(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		response := w.Get(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}

func (w *HTTPRequest) AsyncPostJson(payload []byte) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		response := w.PostJson(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}

func (w *HTTPRequest) AsyncGetJson(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}
		response := w.GetJson(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}

func (w *HTTPRequest) AsyncPost(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		response := w.Post(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}
