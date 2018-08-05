package dnoxy

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/miekg/dns"
)

type HTTPExchangerOptions struct{}

func NewHTTPExchanger(url string, opts *HTTPExchangerOptions) (*HTTPExchanger, error) {
	return &HTTPExchanger{
		url:    url,
		client: &http.Client{},
		opts:   opts,
	}, nil
}

// TODO remove after dev
var _ Exchanger = &HTTPExchanger{}

type HTTPExchanger struct {
	url    string
	client *http.Client
	opts   *HTTPExchangerOptions
}

func (h *HTTPExchanger) Exchange(ctx context.Context, m *dns.Msg) (*dns.Msg, error) {
	b, err := m.Pack()
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, h.url, payload)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("Accept", "application/dns-message")
	req.Header.Add("Content-Type", "application/dns-message")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %v", resp.Status)
	}

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := new(dns.Msg)
	if err := r.Unpack(rb); err != nil {
		return nil, err
	}
	return r, nil
}
