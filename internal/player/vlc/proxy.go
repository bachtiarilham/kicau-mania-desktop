package vlc

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// VLC does not expose arbitrary HTTP request headers. Keep the engine nonce in
// Go and serve only this recording through an unguessable loopback capability.
func recordingProxy(source, headers string) (string, func(), error) {
	u, e := url.Parse(source)
	if e != nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", nil, errors.New("Sumber rekaman tidak valid")
	}
	ip := net.ParseIP(u.Hostname())
	if u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return "", nil, errors.New("Proxy rekaman hanya menerima engine loopback")
	}
	name, value, ok := strings.Cut(headers, ": ")
	if !ok || name != "X-Local-Control" || len(value) < 32 || strings.ContainsAny(value, "\r\n") {
		return "", nil, errors.New("Credential rekaman tidak valid")
	}
	random := make([]byte, 32)
	if _, e = rand.Read(random); e != nil {
		return "", nil, e
	}
	path := "/" + hex.EncodeToString(random)
	proxy := &httputil.ReverseProxy{Director: func(r *http.Request) {
		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.URL.Path = u.Path
		r.URL.RawPath = ""
		r.URL.RawQuery = ""
		r.Host = u.Host
		byteRange := r.Header.Get("Range")
		ifRange := r.Header.Get("If-Range")
		r.Header = make(http.Header)
		r.Header.Set(name, value)
		if byteRange != "" {
			r.Header.Set("Range", byteRange)
		}
		if ifRange != "" {
			r.Header.Set("If-Range", ifRange)
		}
	}, ErrorHandler: func(w http.ResponseWriter, r *http.Request, e error) { http.Error(w, "Rekaman tidak tersedia", 502) }}
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		return "", nil, e
	}
	server := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path || r.URL.RawQuery != "" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(405)
			return
		}
		proxy.ServeHTTP(w, r)
	})}
	go func() { _ = server.Serve(listener) }()
	return "http://" + listener.Addr().String() + path, func() { _ = server.Close() }, nil
}
