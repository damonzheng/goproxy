package httpproxy

import (
	"io"
	"net/http"

	"./filters"

	"github.com/golang/glog"
)

type Handler struct {
	http.Handler
	Listener         Listener
	RequestFilters   []filters.RequestFilter
	RoundTripFilters []filters.RoundTripFilter
	ResponseFilters  []filters.ResponseFilter
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var err error

	remoteAddr := req.RemoteAddr

	// Prepare filter.Context
	ctx := filters.NewContext(h.Listener, rw, req)

	// Enable transport http proxy
	if req.Method != "CONNECT" && !req.URL.IsAbs() {
		if req.TLS != nil {
			req.URL.Scheme = "https"
			if req.Host != "" {
				req.URL.Host = req.Host
			} else {
				req.URL.Host = req.TLS.ServerName
			}
		} else {
			req.URL.Scheme = "http"
			req.URL.Host = req.Host
		}
	}

	// Filter Request
	for _, f := range h.RequestFilters {
		ctx, req, err = f.Request(ctx, req)
		// A roundtrip filter hijacked
		if ctx.Hijacked() {
			return
		}
		if err != nil {
			if err != io.EOF {
				glog.Errorf("%s Filter Request %T(%v) error: %v", remoteAddr, f, f, err)
			}
			return
		}
	}

	// Filter Request -> Response
	var resp *http.Response
	for _, f := range h.RoundTripFilters {
		ctx, resp, err = f.RoundTrip(ctx, req)
		// A roundtrip filter hijacked
		if ctx.Hijacked() {
			return
		}
		// Unexcepted errors
		if err != nil {
			glog.Errorf("%s Filter RoundTrip %T(%v) error: %v", remoteAddr, f, f, err)
			return
		}
		// A roundtrip filter give a response
		if resp != nil {
			resp.Request = req
			break
		}
	}

	// Filter Response
	for _, f := range h.ResponseFilters {
		if resp == nil {
			return
		}
		ctx, resp, err = f.Response(ctx, resp)
		if err != nil {
			glog.Errorf("%s Filter Response %T(%v) error: %v", remoteAddr, f, f, err)
			return
		}
	}

	if resp == nil {
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			rw.Header().Add(key, value)
		}
	}
	rw.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		defer resp.Body.Close()
		n, err := IoCopy(rw, resp.Body)
		if err != nil {
			glog.Errorf("IoCopy %#v return %#v %s", resp.Body, n, err)
		}
	}
}
