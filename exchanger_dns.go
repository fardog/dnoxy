package dnoxy

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
)

type DNSExchangerOptions struct{}

func NewDNSExchanger(addresses []string, opts *DNSExchangerOptions) (*DNSExchanger, error) {
	return &DNSExchanger{
		addresses: addresses,
		client:    new(dns.Client),
		opts:      opts,
	}, nil
}

// TODO remove after dev
var _ Exchanger = &DNSExchanger{}

type DNSExchanger struct {
	addresses []string
	client    *dns.Client
	opts      *DNSExchangerOptions
}

func (d *DNSExchanger) Exchange(ctx context.Context, m *dns.Msg) (r *dns.Msg, err error) {
	for _, a := range d.addresses {
		r, _, err = d.client.Exchange(m, a)
		if err == nil {
			return r, nil
		}
	}

	return nil, fmt.Errorf("unable to reach dns servers; last error was: %v", err)
}
