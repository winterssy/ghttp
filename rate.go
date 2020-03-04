package ghttp

import "golang.org/x/time/rate"

type (
	rateLimiter struct {
		base *rate.Limiter
	}

	concurrency struct {
		ch chan struct{}
	}
)

// Enter implements BeforeRequestCallback interface.
func (rl *rateLimiter) Enter(req *Request) error {
	return rl.base.Wait(req.Context())
}

// Enter implements BeforeRequestCallback interface.
func (c *concurrency) Enter(req *Request) error {
	select {
	case <-req.Context().Done():
		return req.Context().Err()
	case c.ch <- struct{}{}:
		return nil
	}
}

// Exit implements AfterResponseCallback interface.
func (c *concurrency) Exit(*Response, error) {
	<-c.ch
}
