package ghttp

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

type (
	clientTrace struct {
		start                time.Time
		dnsStart             time.Time
		dnsDone              time.Time
		connStart            time.Time
		connDone             time.Time
		tlsHandshakeStart    time.Time
		tlsHandshakeDone     time.Time
		gotFirstResponseByte time.Time
		wroteRequest         time.Time
		end                  time.Time
		getConn              time.Time
		gotConn              time.Time
		gotConnInfo          httptrace.GotConnInfo
	}

	// TraceInfo is used to provide request trace info such as DNS lookup
	// duration, Connection obtain duration, Server processing duration, etc.
	TraceInfo struct {
		// DNSLookupTime is a duration that transport took to perform
		// DNS lookup.
		DNSLookupTime time.Duration `json:"dns_lookup_time"`

		// TCPConnTime is a duration that TCP connection took place.
		TCPConnTime time.Duration `json:"tcp_conn_time"`

		// TLSHandshakeTime is a duration that TLS handshake took place.
		TLSHandshakeTime time.Duration `json:"tls_handshake_time,omitempty"`

		// ConnTime is a duration that took to obtain a successful connection.
		ConnTime time.Duration `json:"conn_time"`

		// ServerTime is a duration that server took to respond first byte.
		ServerTime time.Duration `json:"server_time"`

		// ResponseTime is a duration since first response byte from server to
		// request completion.
		ResponseTime time.Duration `json:"response_time"`

		// TotalTime is a duration that total request took end-to-end.
		TotalTime time.Duration `json:"total_time"`

		// ConnReused reports whether this connection has been previously
		// used for another HTTP request.
		ConnReused bool `json:"conn_reused"`

		// ConnWasIdle reports whether this connection was obtained from an
		// idle pool.
		ConnWasIdle bool `json:"conn_was_idle"`

		// ConnIdleTime is a duration how long the connection was previously
		// idle, if ConnWasIdle is true.
		ConnIdleTime time.Duration `json:"conn_idle_time"`
	}
)

func (ct *clientTrace) modifyRequest(req *Request) {
	ctx := httptrace.WithClientTrace(
		req.Context(),
		&httptrace.ClientTrace{
			GetConn: func(_ string) {
				ct.getConn = time.Now()
			},
			GotConn: func(gotConnInfo httptrace.GotConnInfo) {
				ct.gotConn = time.Now()
				ct.gotConnInfo = gotConnInfo
			},
			GotFirstResponseByte: func() {
				ct.gotFirstResponseByte = time.Now()
			},
			DNSStart: func(_ httptrace.DNSStartInfo) {
				ct.dnsStart = time.Now()
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				ct.dnsDone = time.Now()
			},
			ConnectStart: func(network, addr string) {
				ct.connStart = time.Now()
			},
			ConnectDone: func(network, addr string, err error) {
				ct.connDone = time.Now()
			},
			TLSHandshakeStart: func() {
				ct.tlsHandshakeStart = time.Now()
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
				ct.tlsHandshakeDone = time.Now()
			},
			WroteRequest: func(_ httptrace.WroteRequestInfo) {
				ct.wroteRequest = time.Now()
			},
		},
	)
	req.Request = req.WithContext(ctx)
}

func (ct *clientTrace) done() {
	ct.end = time.Now()
}

func (ct *clientTrace) traceInfo() *TraceInfo {
	return &TraceInfo{
		DNSLookupTime:    ct.dnsDone.Sub(ct.dnsStart),
		TCPConnTime:      ct.connDone.Sub(ct.connStart),
		TLSHandshakeTime: ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart),
		ConnTime:         ct.gotConn.Sub(ct.getConn),
		ServerTime:       ct.gotFirstResponseByte.Sub(ct.wroteRequest),
		ResponseTime:     ct.end.Sub(ct.gotFirstResponseByte),
		TotalTime:        ct.end.Sub(ct.start),
		ConnReused:       ct.gotConnInfo.Reused,
		ConnWasIdle:      ct.gotConnInfo.WasIdle,
		ConnIdleTime:     ct.gotConnInfo.IdleTime,
	}
}
