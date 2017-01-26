package spotify

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type rateLimitAwareTransport struct {
	proxied http.RoundTripper
}

func (rlt *rateLimitAwareTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := cloneRequest(r)
	resp, err := rlt.proxied.RoundTrip(r2)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		return resp, nil
	}

	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return nil, fmt.Errorf("API Rate Limit reached but Retry-After header is not set")
	}

	sleep, err := time.ParseDuration(retryAfter + "s")
	if err != nil {
		return nil, fmt.Errorf("Unable to parse value of Retry-After header into Duration: %s", err)
	}

	log.Printf("Sleeping for %s seconds\n", retryAfter)
	time.Sleep(sleep)

	return rlt.proxied.RoundTrip(r2)
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
