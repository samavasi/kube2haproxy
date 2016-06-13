package flowcontrol

import (
	"sync"

	"github.com/juju/ratelimit"
)

type RateLimiter interface {
	// TryAccept returns true if a token is taken immediately. Otherwise,
	// it returns false.
	TryAccept() bool
	// Accept returns once a token becomes available.
	Accept()
	// Stop stops the rate limiter, subsequent calls to CanAccept will return false
	Stop()
	// Saturation returns a percentage number which describes how saturated
	// this rate limiter is.
	// Usually we use token bucket rate limiter. In that case,
	// 1.0 means no tokens are available; 0.0 means we have a full bucket of tokens to use.
	Saturation() float64
}

type tokenBucketRateLimiter struct {
	limiter *ratelimit.Bucket
}

// NewTokenBucketRateLimiter creates a rate limiter which implements a token bucket approach.
// The rate limiter allows bursts of up to 'burst' to exceed the QPS, while still maintaining a
// smoothed qps rate of 'qps'.
// The bucket is initially filled with 'burst' tokens, and refills at a rate of 'qps'.
// The maximum number of tokens in the bucket is capped at 'burst'.
func NewTokenBucketRateLimiter(qps float32, burst int) RateLimiter {
	limiter := ratelimit.NewBucketWithRate(float64(qps), int64(burst))
	return &tokenBucketRateLimiter{limiter}
}

func (t *tokenBucketRateLimiter) TryAccept() bool {
	return t.limiter.TakeAvailable(1) == 1
}

func (t *tokenBucketRateLimiter) Saturation() float64 {
	capacity := t.limiter.Capacity()
	avail := t.limiter.Available()
	return float64(capacity-avail) / float64(capacity)
}

// Accept will block until a token becomes available
func (t *tokenBucketRateLimiter) Accept() {
	t.limiter.Wait(1)
}

func (t *tokenBucketRateLimiter) Stop() {
}

type fakeAlwaysRateLimiter struct{}

func NewFakeAlwaysRateLimiter() RateLimiter {
	return &fakeAlwaysRateLimiter{}
}

func (t *fakeAlwaysRateLimiter) TryAccept() bool {
	return true
}

func (t *fakeAlwaysRateLimiter) Saturation() float64 {
	return 0
}

func (t *fakeAlwaysRateLimiter) Stop() {}

func (t *fakeAlwaysRateLimiter) Accept() {}

type fakeNeverRateLimiter struct {
	wg sync.WaitGroup
}

func NewFakeNeverRateLimiter() RateLimiter {
	wg := sync.WaitGroup{}
	wg.Add(1)
	return &fakeNeverRateLimiter{
		wg: wg,
	}
}

func (t *fakeNeverRateLimiter) TryAccept() bool {
	return false
}

func (t *fakeNeverRateLimiter) Saturation() float64 {
	return 1
}

func (t *fakeNeverRateLimiter) Stop() {
	t.wg.Done()
}

func (t *fakeNeverRateLimiter) Accept() {
	t.wg.Wait()
}