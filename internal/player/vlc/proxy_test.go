package vlc

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecordingProxyKeepsNonceInGoAndPreservesRanges(t *testing.T) {
	nonce := strings.Repeat("n", 64)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/recordings/rec/file" || r.Header.Get("X-Local-Control") != nonce || r.Header.Get("Range") != "bytes=0-2" || r.Header.Get("Cookie") != "" {
			t.Error("incorrect authenticated recording request")
		}
		w.Header().Set("Content-Type", "video/x-matroska")
		w.Header().Set("Content-Range", "bytes 0-2/3")
		w.WriteHeader(206)
		_, _ = w.Write([]byte("mkv"))
	}))
	defer upstream.Close()
	address, stop, e := recordingProxy(upstream.URL+"/api/v1/recordings/rec/file", "X-Local-Control: "+nonce)
	if e != nil {
		t.Fatal(e)
	}
	defer stop()
	if strings.Contains(address, nonce) {
		t.Fatal("nonce exposed to VLC URL")
	}
	req, _ := http.NewRequest("GET", address, nil)
	req.Header.Set("Range", "bytes=0-2")
	req.Header.Set("Cookie", "untrusted=value")
	res, e := http.DefaultClient.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != 206 || string(b) != "mkv" {
		t.Fatal("range playback broken")
	}
	res, e = http.Get(address + "/other")
	if e != nil {
		t.Fatal(e)
	}
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Fatal("proxy allowed another path")
	}
	stop()
	if _, e = http.Get(address); e == nil {
		t.Fatal("proxy survived player shutdown")
	}
}
func TestRecordingProxyRejectsExternalCredentialTargets(t *testing.T) {
	for _, source := range []string{"https://example.com/file", "http://192.0.2.1/file", "http://user:password@127.0.0.1/file"} {
		if _, stop, e := recordingProxy(source, "X-Local-Control: "+strings.Repeat("n", 64)); e == nil {
			stop()
			t.Fatal("external source accepted")
		}
	}
}
