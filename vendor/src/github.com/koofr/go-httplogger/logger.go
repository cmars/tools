// Package httplogger provides http.RoundTripper wrapper for debugging HTTP client.
package httplogger

import (
	"io"
	"net/http"
	"net/http/httputil"
	"os"
)

// TransportLogger is an implementation of RoundTripper that takes
// another RoundTripper and loggs all traffic to the Writer.
type TransportLogger struct {
	Transport http.RoundTripper
	Writer    io.Writer
}

// New creates new TransportLogger for given transport.
func New(transport http.RoundTripper) *TransportLogger {
	return &TransportLogger{
		Transport: transport,
	}
}

// NewDefault creates new TransportLogger for http.DefaultTransport
func NewDefault() *TransportLogger {
	return &TransportLogger{}
}

// RoundTrip implements the RoundTripper interface.
func (t *TransportLogger) RoundTrip(req *http.Request) (res *http.Response, err error) {
	transport := t.Transport

	if transport == nil {
		transport = http.DefaultTransport
	}

	writer := t.Writer

	if writer == nil {
		writer = os.Stdout
	}

	reqDump, err := httputil.DumpRequestOut(req, true)

	writer.Write(reqDump)
	writer.Write([]byte("\n"))

	res, err = transport.RoundTrip(req)

	if err != nil {
		return
	}

	resDump, err := httputil.DumpResponse(res, true)

	writer.Write(resDump)
	writer.Write([]byte("\n"))

	return
}
