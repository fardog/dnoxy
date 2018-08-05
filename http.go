package dnoxy

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type HTTPHandlerOptions struct{}

func NewHTTPHandler(ex Exchanger, opts *HTTPHandlerOptions) (*HTTPHandler, error) {
	return &HTTPHandler{
		ex: ex,
	}, nil
}

type HTTPHandler struct {
	ex Exchanger
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	status := http.StatusInternalServerError
	defer func() {
		if err != nil {
			log.WithField("status", status).Errorf(err.Error())
			w.WriteHeader(status)
			w.Write([]byte(err.Error()))
		}
	}()

	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		err = errors.New("method unsupported")
		return
	}

	if r.Header.Get("Content-Type") != "application/dns-message" {
		status = http.StatusBadRequest
		err = errors.New("invalid content-type")
		return
	}
	if r.Header.Get("Accept") != "application/dns-message" {
		status = http.StatusNotAcceptable
		err = errors.New("invalid accept header")
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	msg := new(dns.Msg)
	err = msg.Unpack(b)
	if err != nil {
		return
	}

	resp, err := h.ex.Exchange(r.Context(), msg)
	if err != nil {
		return
	}

	rb, err := resp.Pack()
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(rb)
	if err != nil {
		return
	}

	log.WithFields(log.Fields{
		"status":   200,
		"question": msg.Question[0].String(),
	}).Infof("responding")
}
