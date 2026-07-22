package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	requestTimeout   = 15 * time.Second
	maxResponseBytes = 32 << 20
)

// Client is the narrow metadata contract required by the scheduler.
type Client interface {
	GetVersion() (string, error)
	GetHosts() ([]Host, error)
	GetContainers() ([]Container, error)
	OnChangeWithError(int, func(string)) error
}

type httpClient struct {
	baseURL *url.URL
	client  *http.Client
	initErr error
}

// NewClient preserves the historical constructor shape while validating the
// endpoint before the first request.
func NewClient(rawURL string) Client {
	client, err := newHTTPClient(rawURL)
	if err != nil {
		return &httpClient{initErr: err}
	}
	return client
}

func newHTTPClient(rawURL string) (*httpClient, error) {
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.New("invalid metadata URL")
	}
	if (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" {
		return nil, errors.New("metadata URL must use HTTP or HTTPS and include a host")
	}
	if baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, errors.New("metadata URL must not contain credentials, a query, or a fragment")
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")

	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Metadata version requests are long polls. Keep the transport deadline at
	// least as long as the per-request deadline so a valid long poll is not
	// aborted before its context expires.
	transport.ResponseHeaderTimeout = requestTimeout
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.IdleConnTimeout = 30 * time.Second

	httpTransport := &http.Client{Transport: transport}
	httpTransport.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) == 0 {
			return nil
		}
		origin := via[0].URL
		if !strings.EqualFold(request.URL.Scheme, origin.Scheme) || !strings.EqualFold(request.URL.Host, origin.Host) {
			return errors.New("metadata redirect changed origin")
		}
		if len(via) >= 5 {
			return errors.New("too many metadata redirects")
		}
		return nil
	}

	return &httpClient{baseURL: baseURL, client: httpTransport}, nil
}

func (client *httpClient) endpoint(path string) (*url.URL, error) {
	if client.initErr != nil {
		return nil, client.initErr
	}
	reference, err := url.Parse(path)
	if err != nil || reference.IsAbs() || reference.Host != "" || !strings.HasPrefix(reference.Path, "/") {
		return nil, errors.New("invalid metadata request path")
	}
	for _, segment := range strings.Split(reference.EscapedPath(), "/") {
		if segment == ".." || strings.EqualFold(segment, "%2e%2e") {
			return nil, errors.New("metadata request path contains traversal")
		}
	}
	target := *client.baseURL
	target.Path = client.baseURL.Path + reference.Path
	target.RawPath = ""
	target.RawQuery = reference.RawQuery
	return &target, nil
}

func (client *httpClient) sendRequest(path string, timeout time.Duration) ([]byte, error) {
	target, err := client.endpoint(path)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, errors.New("could not create metadata request")
	}
	request.Header.Set("Accept", "application/json")
	response, err := client.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("metadata request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata request returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return nil, errors.New("could not read metadata response")
	}
	if len(body) > maxResponseBytes {
		return nil, errors.New("metadata response exceeded size limit")
	}
	return body, nil
}

func (client *httpClient) decode(path string, destination interface{}) error {
	body, err := client.sendRequest(path, requestTimeout)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, destination); err != nil {
		return errors.New("metadata response was not valid JSON")
	}
	return nil
}

func (client *httpClient) GetVersion() (string, error) {
	body, err := client.sendRequest("/version", requestTimeout)
	if err != nil {
		return "", err
	}
	var version string
	if err := json.Unmarshal(body, &version); err == nil && version != "" {
		return version, nil
	}
	version = strings.TrimSpace(string(body))
	if version == "" {
		return "", errors.New("metadata version was empty")
	}
	return version, nil
}

func (client *httpClient) GetHosts() ([]Host, error) {
	var result []Host
	return result, client.decode("/hosts", &result)
}

func (client *httpClient) GetContainers() ([]Container, error) {
	var result []Container
	return result, client.decode("/containers", &result)
}

func (client *httpClient) waitVersion(maxWait int, version string) (string, error) {
	if maxWait < 1 {
		maxWait = 1
	}
	values := url.Values{}
	values.Set("wait", "true")
	values.Set("value", version)
	values.Set("maxWait", strconv.Itoa(maxWait))
	for {
		body, err := client.sendRequest("/version?"+values.Encode(), time.Duration(maxWait+10)*time.Second)
		if err != nil {
			var timeout interface {
				Timeout() bool
			}
			if errors.As(err, &timeout) && timeout.Timeout() {
				continue
			}
			return "", err
		}
		var next string
		if err := json.Unmarshal(body, &next); err == nil && next != "" {
			return next, nil
		}
		next = strings.TrimSpace(string(body))
		if next == "" {
			return "", errors.New("metadata version was empty")
		}
		return next, nil
	}
}

func (client *httpClient) OnChangeWithError(intervalSeconds int, callback func(string)) error {
	if callback == nil {
		return errors.New("metadata change callback is nil")
	}
	version := "init"
	for {
		next, err := client.waitVersion(intervalSeconds, version)
		if err != nil {
			return err
		}
		if next != version {
			version = next
			callback(next)
		}
	}
}
