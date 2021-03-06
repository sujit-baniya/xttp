package xttp

import (
	"net/http"
	"time"
)

// Client is used to make HTTP requests. It adds additional functionality
// like automatic retries to tolerate minor outages.
type Client struct {
	// HTTPClient is the internal HTTP client.
	HTTPClient *http.Client

	// RequestLogHook allows a user-supplied function to be called
	// before each retry.
	RequestLogHook RequestLogHook
	// ResponseLogHook allows a user-supplied function to be called
	// with the response from each HTTP request executed.
	ResponseLogHook ResponseLogHook
	// ErrorHandler specifies the custom error handler to use, if any
	ErrorHandler ErrorHandler

	// CheckRetry specifies the policy for handling retries, and is called
	// after each request. The default policy is DefaultRetryPolicy.
	CheckRetry CheckRetry
	// Backoff specifies the policy for how long to wait between retries
	Backoff Backoff

	options Options
}

// Options contains configuration options for the client
type Options struct {
	// RetryWaitMin is the minimum time to wait for retry
	RetryWaitMin time.Duration
	// RetryWaitMax is the maximum time to wait for retry
	RetryWaitMax time.Duration
	// Timeout is the maximum time to wait for the request
	Timeout time.Duration
	// RetryMax is the maximum number of retries
	RetryMax int
	// RespReadLimit is the maximum HTTP response size to read for
	// connection being reused.
	RespReadLimit int64
	// Verbose specifies if debug messages should be printed
	Verbose bool
	// KillIdleConn specifies if all keep-alive connections gets killed
	KillIdleConn bool
	MaxPoolSize  int
	ReqPerSec    int
	Semaphore    chan int
	RateLimiter  <-chan time.Time
}

// DefaultOptionsSpraying contains the default options for host spraying
// scenarios where lots of requests need to be sent to different hosts.
var DefaultOptionsSpraying = Options{
	RetryWaitMin:  1 * time.Second,
	RetryWaitMax:  30 * time.Second,
	Timeout:       30 * time.Second,
	RetryMax:      5,
	RespReadLimit: 4096,
	KillIdleConn:  true,
	MaxPoolSize:   100,
	ReqPerSec:     10,
}

// DefaultOptionsSingle contains the default options for host bruteforce
// scenarios where lots of requests need to be sent to a single host.
var DefaultOptionsSingle = Options{
	RetryWaitMin:  1 * time.Second,
	RetryWaitMax:  30 * time.Second,
	Timeout:       30 * time.Second,
	RetryMax:      5,
	RespReadLimit: 4096,
	KillIdleConn:  false,
	MaxPoolSize:   100,
	ReqPerSec:     10,
}

// NewClient creates a new Client with default settings.
func NewClient(options Options) *Client {
	httpclient := DefaultClient()
	// if necessary adjusts per-request timeout proportionally to general timeout (30%)
	if options.Timeout > time.Second*15 {
		httpclient.Timeout = time.Duration(options.Timeout.Seconds()*0.3) * time.Second
	}
	var semaphore chan int = nil
	if options.MaxPoolSize > 0 {
		semaphore = make(chan int, options.MaxPoolSize) // Buffered channel to act as a semaphore
	}

	var emitter <-chan time.Time = nil
	if options.ReqPerSec > 0 {
		emitter = time.NewTicker(time.Second / time.Duration(options.ReqPerSec)).C // x req/s == 1s/x req (inverse)
	}
	options.Semaphore = semaphore
	options.RateLimiter = emitter

	c := &Client{
		HTTPClient: httpclient,
		CheckRetry: DefaultRetryPolicy(),
		Backoff:    ExponentialJitterBackoff(),
		options:    options,
	}

	c.setKillIdleConnections()
	return c
}

// NewWithHTTPClient creates a new Client with default settings and provided hclient.Client
func NewWithHTTPClient(client *http.Client, options Options) *Client {
	c := &Client{
		HTTPClient: client,
		CheckRetry: DefaultRetryPolicy(),
		Backoff:    DefaultBackoff(),

		options: options,
	}

	c.setKillIdleConnections()
	return c
}

// setKillIdleConnections sets the kill idle conns switch in two scenarios
//  1. If the hclient.Client has settings that require us to do so.
//  2. The user has enabled it by default, in which case we have nothing to do.
func (c *Client) setKillIdleConnections() {
	if c.HTTPClient != nil || !c.options.KillIdleConn {
		if b, ok := c.HTTPClient.Transport.(*http.Transport); ok {
			c.options.KillIdleConn = b.DisableKeepAlives || b.MaxConnsPerHost < 0
		}
	}
}
