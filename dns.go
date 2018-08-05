package dnoxy

import (
	"context"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type DNSHandlerOptions struct{}

func NewDNSHandler(ex Exchanger, opts *DNSHandlerOptions) (*DNSHandler, error) {
	return &DNSHandler{
		ex: ex,
	}, nil
}

type DNSHandler struct {
	ex Exchanger
}

func (h *DNSHandler) Handle(w dns.ResponseWriter, r *dns.Msg) {
	resp, err := h.ex.Exchange(context.Background(), r)
	if err != nil {
		log.Error(err)
		dns.HandleFailed(w, r)
		return
	}

	if err := w.WriteMsg(resp); err != nil {
		log.Error(err)
		dns.HandleFailed(w, r)
		return
	}

	log.WithFields(log.Fields{
		"question": r.Question[0].String(),
	}).Infof("responded")
}
