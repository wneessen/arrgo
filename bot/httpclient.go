package bot

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// HTTPClient is an object wrapper for the Go http.Client
type HTTPClient struct {
	*http.Client
}

// HTTPRequest is an object wrapper for the Go http.Request
type HTTPRequest struct {
	*http.Request
}

// HTTPReqMethod is a wrapper around a string
type HTTPReqMethod string

// List of HTTPReqMethods
const (
	// ReqMethodGet is the GET request method
	ReqMethodGet HTTPReqMethod = "GET"

	// ReqMethodPost is the POST request method
	ReqMethodPost HTTPReqMethod = "POST"
)

// SOTReferer is the referer that apparently is needed for the SoT API to accept requests
const SOTReferer = "https://www.seaofthieves.com/profile/achievements"

// NewHTTPClient returns a HTTPClient object
func NewHTTPClient() (*HTTPClient, error) {
	tc := &tls.Config{
		MaxVersion:    tls.VersionTLS13,
		MinVersion:    tls.VersionTLS12,
		Renegotiation: tls.RenegotiateFreelyAsClient,
	}
	t := &http.Transport{TLSClientConfig: tc}
	cj, err := cookiejar.New(nil)
	if err != nil {
		return &HTTPClient{}, err
	}
	hc := &http.Client{
		Transport: t,
		Timeout:   10 * time.Second,
		Jar:       cj,
	}

	return &HTTPClient{hc}, nil
}

// HttpReq generates a HTTPRequest based on the Request method and request URI
func (h *HTTPClient) HttpReq(p string, m HTTPReqMethod, q map[string]string) (*HTTPRequest, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, err
	}

	if m == http.MethodGet {
		uq := u.Query()
		for k, v := range q {
			uq.Add(k, v)
		}
		u.RawQuery = uq.Encode()
	}

	hr, err := http.NewRequest(string(m), u.String(), nil)
	if err != nil {
		return nil, err
	}

	if m == http.MethodPost {
		pd := url.Values{}
		for k, v := range q {
			pd.Add(k, v)
		}

		rb := io.NopCloser(bytes.NewBufferString(pd.Encode()))
		hr.Body = rb
	}
	hr.Header.Set("user-agent", fmt.Sprintf(`ArrGo Bot v%s (https://www.github.com/wneessen/arrgo)`,
		Version))

	return &HTTPRequest{hr}, nil
}

// Fetch performs the actual HTTP request
func (h *HTTPClient) Fetch(r *HTTPRequest) ([]byte, *http.Response, error) {
	res, err := h.Do(r.Request)
	if err != nil {
		return nil, res, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	hb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res, err
	}
	return hb, res, nil
}

// SetSOTRequest sets the required additional headers for Sea of Thieves API requests
func (r *HTTPRequest) SetSOTRequest(c string) {
	r.Header.Set("referer", SOTReferer)
	rc := &http.Cookie{Name: "rat", Value: c}
	r.AddCookie(rc)
}

// SetReferer sets a custom referer to the request
func (r *HTTPRequest) SetReferer(rf string) {
	r.Header.Set("referer", rf)
}
