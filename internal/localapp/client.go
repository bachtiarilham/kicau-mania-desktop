package localapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	mu    sync.Mutex
	Base  string
	nonce string
	http  *http.Client
}

func New(base string) *Client {
	return &Client{Base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (c *Client) Bootstrap(ctx context.Context) (map[string]any, error) {
	var r map[string]any
	if e := c.Call(ctx, "GET", "/api/v1/setup/bootstrap", nil, &r); e != nil {
		return nil, e
	}
	nonce, _ := r["nonce"].(string)
	if len(nonce) < 32 {
		return nil, errors.New("Nonce engine tidak valid")
	}
	c.mu.Lock()
	c.nonce = nonce
	c.mu.Unlock()
	delete(r, "nonce")
	return r, nil
}
func (c *Client) Nonce() string { c.mu.Lock(); defer c.mu.Unlock(); return c.nonce }
func (c *Client) Call(ctx context.Context, method, path string, body, out any) error {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") || strings.ContainsAny(path, "\\#") {
		return errors.New("Path engine tidak valid")
	}
	var raw io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return e
		}
		raw = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(ctx, method, c.Base+path, raw)
	if e != nil {
		return e
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Local-Control", c.Nonce())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, e := c.http.Do(req)
	if e != nil {
		return errors.New("Local-app tidak dapat dihubungi")
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var env struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&env)
		if env.Message == "" {
			env.Message = "Permintaan engine ditolak"
		}
		return fmt.Errorf("%s (%d)", env.Message, res.StatusCode)
	}
	if res.StatusCode == 204 {
		return nil
	}
	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if e = json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&env); e != nil || !env.Success {
		return errors.New("Respons local-app tidak valid")
	}
	if out == nil {
		return nil
	}
	if len(env.Data) == 0 {
		env.Data = json.RawMessage("null")
	}
	return json.Unmarshal(env.Data, out)
}
func (c *Client) File(ctx context.Context, id string) (*http.Response, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", c.Base+"/api/v1/recordings/"+id+"/file", nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("X-Local-Control", c.Nonce())
	client := *c.http
	client.Timeout = 0
	r, e := client.Do(req)
	if e != nil {
		return nil, errors.New("Rekaman lokal tidak dapat dibaca")
	}
	if r.StatusCode != 200 {
		r.Body.Close()
		return nil, errors.New("Rekaman lokal belum tersedia")
	}
	return r, nil
}
func (c *Client) Metrics(ctx context.Context, emit func(any)) {
	for ctx.Err() == nil {
		url := strings.Replace(strings.Replace(c.Base, "https://", "wss://", 1), "http://", "ws://", 1) + "/api/v1/metrics"
		dial := websocket.Dialer{Subprotocols: []string{"edge-v1"}, HandshakeTimeout: 5 * time.Second}
		conn, _, e := dial.DialContext(ctx, url, nil)
		if e == nil {
			emit(map[string]any{"connection": "connected"})
			done := make(chan struct{})
			go func() {
				select {
				case <-ctx.Done():
					conn.Close()
				case <-done:
				}
			}()
			for {
				var frame map[string]any
				if e = conn.ReadJSON(&frame); e != nil {
					break
				}
				emit(frame)
			}
			close(done)
			conn.Close()
		}
		emit(map[string]any{"connection": "disconnected"})
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}
