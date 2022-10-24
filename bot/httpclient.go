package bot

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wneessen/arrgo/crypto"
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

// APIIntString represents a API response string that is actually a Integer
type APIIntString int64

// APITimeRFC3339 is a wrapper type for time.Time
type APITimeRFC3339 time.Time

// SOTReferer is the referer that apparently is needed for the SoT API to accept requests
const SOTReferer = "https://www.seaofthieves.com/profile/achievements"

// HTTP client related errors
var (
	// ErrSOTUnauth should be used when requrests to the SoT API were not successful due to expired
	// tokens
	ErrSOTUnauth = errors.New("failed to fetch Sea of Thieves content, due to being unauthorized")
)

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
		Timeout:   20 * time.Second,
		Jar:       cj,
	}

	return &HTTPClient{hc}, nil
}

// HTTPReq generates a HTTPRequest based on the Request method and request URI
func (h *HTTPClient) HTTPReq(p string, m HTTPReqMethod, q map[string]string) (*HTTPRequest, error) {
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
	r.SetReferer(SOTReferer)
	r.Header.Set("accept", "application/json")
	r.Header.Set("cache-control", "max-age=0")
	rc := &http.Cookie{Name: "rat", Value: c}
	r.AddCookie(rc)
	rn, err := crypto.RandNum(10000)
	if err == nil {
		r.URL.RawQuery = fmt.Sprintf("x=%d", rn)
	}
}

// SetReferer sets a custom referer to the request
func (r *HTTPRequest) SetReferer(rf string) {
	r.Header.Set("referer", rf)
}

// UnmarshalJSON converts the APIIntString string into an int64
func (s *APIIntString) UnmarshalJSON(ib []byte) error {
	is := string(ib)
	if is == "null" {
		return nil
	}
	is = strings.ReplaceAll(is, `"`, ``)
	realInt, err := strconv.ParseInt(is, 10, 64)
	if err != nil {
		return fmt.Errorf("string to int conversion failed: %w", err)
	}
	*(*int64)(s) = realInt

	return nil
}

// UnmarshalJSON converts a API RFC3339 formated date strings into a
// time.Time object
func (t *APITimeRFC3339) UnmarshalJSON(s []byte) error {
	dateString := string(s)
	dateString = strings.ReplaceAll(dateString, `"`, "")
	if dateString == "null" {
		return nil
	}
	dateParse, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return fmt.Errorf("failed to parse string as RFC3339 time string: %w", err)
	}

	*(*time.Time)(t) = dateParse
	return nil
}
