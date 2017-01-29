package spotify

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
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

	headerVal := resp.Header.Get("Retry-After")
	if headerVal == "" {
		return nil, fmt.Errorf("API Rate Limit reached but Retry-After header is not set")
	}

	retryAfter, err := strconv.ParseInt(headerVal, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Unable to convert value of Retry-After header to int: %s", err)
	}

	var sleep time.Duration
	if retryAfter == 0 {
		sleep = 2 * time.Second
	} else {
		sleep = time.Duration(retryAfter) * time.Second * 2
	}

	log.Printf("Sleeping for %s\n", sleep.String())
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
