package dnoxy

import (
	"context"

	"github.com/miekg/dns"
)

// Exchanger is an interface describing a DNS client over any transport.
type Exchanger interface {
	Exchange(ctx context.Context, m *dns.Msg) (r *dns.Msg, err error)
}
