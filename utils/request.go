package utils

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type RestyClient struct {
	client      *resty.Client
	concurrency int
	sem         chan struct{}
}

func NewRestyClient(concurrency int) *RestyClient {
	client := &RestyClient{
		client:      resty.New(),
		concurrency: concurrency,
		sem:         make(chan struct{}, concurrency),
	}
	client.client.SetTransport(&http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if addr == "www.bilinovel.com:443" {
				addr = "64.140.161.52:443"
			}
			return (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext(ctx, network, addr)
		},
		TLSHandshakeTimeout: 10 * time.Second,
	})
	client.client.
		OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
			client.sem <- struct{}{}
			return nil
		}).
		OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
			<-client.sem
			return nil
		})
	client.client.SetRetryCount(10).
		SetRetryWaitTime(3 * time.Second).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			if resp.StatusCode() == http.StatusTooManyRequests {
				if retryAfter := resp.Header().Get("Retry-After"); retryAfter != "" {
					if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
						return seconds, nil
					}
					if t, err := http.ParseTime(retryAfter); err == nil {
						return time.Until(t), nil
					}
				}
				return 3 * time.Second, nil
			}
			return 0, nil
		}).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() == http.StatusTooManyRequests
		})

	client.client.SetLogger(disableLogger{}).SetHeader("Accept-Charset", "utf-8").SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
	return client
}

func (c *RestyClient) R() *resty.Request {
	return c.client.R()
}

type disableLogger struct{}

func (d disableLogger) Errorf(string, ...interface{}) {}
func (d disableLogger) Warnf(string, ...interface{})  {}
func (d disableLogger) Debugf(string, ...interface{}) {}
