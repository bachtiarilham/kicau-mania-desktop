package gateway

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"kicau.desktop/cctv/internal/config"
)

type Session struct {
	Base    string    `json:"base"`
	Access  string    `json:"access"`
	Refresh string    `json:"refresh"`
	Cookie  string    `json:"cookie"`
	Expires time.Time `json:"expires"`
	Device  string    `json:"device"`
}

type Client struct {
	mu           sync.Mutex
	Base, Origin string
	session      Session
	store        config.ConfigStore
	http         *http.Client
}

type Failure struct {
	Status  int
	Message string
}

func (e *Failure) Error() string { return e.Message }

const signInRequired = "Sesi cloud tidak aktif. Silakan masuk kembali ke cloud."

func ID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	s := hex.EncodeToString(b[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

func New(base string, store config.ConfigStore) *Client {
	base = strings.TrimRight(base, "/")
	if !strings.HasSuffix(base, "/api/v1") {
		base += "/api/v1"
	}
	u, _ := url.Parse(base)
	c := &Client{Base: base, Origin: u.Scheme + "://" + u.Host, store: store, http: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	_ = store.Load("cloud-session.bin", &c.session, true)
	if c.session.Base != base {
		c.session = Session{Base: base}
	}
	if c.session.Device == "" {
		c.session.Device = ID()
	}
	return c
}

func (c *Client) raw(ctx context.Context, method, path string, body any, out any) (*http.Response, error) {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") || strings.ContainsAny(path, "\\#") {
		return nil, errors.New("Path cloud tidak valid")
	}
	var raw io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return nil, e
		}
		raw = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(ctx, method, c.Base+path, raw)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", c.Origin)
	req.Header.Set("User-Agent", "KicauManiaDesktop/1.0")
	if c.session.Device != "" {
		req.Header.Set("X-Device-ID", c.session.Device)
		req.Header.Set("X-Device-Name", "KicauManiaDesktop")
	}
	if raw != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.session.Access != "" {
		req.Header.Set("Authorization", "Bearer "+c.session.Access)
	}
	if strings.HasPrefix(path, "/auth/") && c.session.Refresh != "" {
		req.AddCookie(&http.Cookie{Name: c.session.Cookie, Value: c.session.Refresh})
	}
	res, e := c.http.Do(req)
	if e != nil {
		return nil, errors.New("Cloud tidak dapat dihubungi; CCTV lokal tetap tersedia")
	}
	defer res.Body.Close()
	var env struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	e = json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&env)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		message := env.Message
		if message == "" || res.StatusCode >= 500 {
			message = "Layanan cloud belum tersedia"
		}
		return res, &Failure{res.StatusCode, message}
	}
	if e != nil || !env.Success {
		return res, errors.New("Respons cloud tidak sesuai kontrak")
	}
	if out == nil {
		return res, nil
	}
	return res, json.Unmarshal(env.Data, out)
}

func (c *Client) tokens(res *http.Response, v struct {
	Access       string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Expires      int            `json:"expires_in"`
	User         map[string]any `json:"user"`
}) error {
	if v.Access == "" || v.Expires < 1 {
		return errors.New("Token cloud tidak valid")
	}
	found := false
	for _, cookie := range res.Cookies() {
		if cookie.Name == "__Host-kicau_refresh" || cookie.Name == "kicau_refresh" {
			if cookie.Value != "" {
				c.session.Refresh = cookie.Value
				c.session.Cookie = cookie.Name
				found = true
			}
		}
	}
	if !found && v.RefreshToken != "" {
		c.session.Refresh = v.RefreshToken
		c.session.Cookie = "kicau_refresh"
		found = true
	}
	if !found {
		return errors.New("Cloud tidak mengirim refresh cookie")
	}
	c.session.Access = v.Access
	c.session.Expires = time.Now().Add(time.Duration(v.Expires) * time.Second)
	return c.store.Save("cloud-session.bin", c.session, true)
}

func (c *Client) refresh(ctx context.Context) error {
	if c.session.Refresh == "" {
		return &Failure{http.StatusUnauthorized, signInRequired}
	}
	var v struct {
		Access       string         `json:"access_token"`
		RefreshToken string         `json:"refresh_token"`
		Expires      int            `json:"expires_in"`
		User         map[string]any `json:"user"`
	}
	r, e := c.raw(ctx, "POST", "/auth/refresh", map[string]any{}, &v)
	if e != nil {
		var f *Failure
		if errors.As(e, &f) && f.Status == http.StatusUnauthorized {
			c.session = Session{Base: c.Base, Device: c.session.Device}
			_ = c.store.Save("cloud-session.bin", c.session, true)
			return &Failure{http.StatusUnauthorized, signInRequired}
		}
		return e
	}
	return c.tokens(r, v)
}

func (c *Client) Login(ctx context.Context, email, password string) (map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var pre struct {
		PreAuthToken     string `json:"pre_auth_token"`
		AccessibleScopes []struct {
			ScopeID   string  `json:"scope_id"`
			Role      string  `json:"role"`
			ScopeType string  `json:"scope_type"`
			ArenaID   *string `json:"arena_id"`
		} `json:"accessible_scopes"`
		Access       string         `json:"access_token"`
		RefreshToken string         `json:"refresh_token"`
		Expires      int            `json:"expires_in"`
		User         map[string]any `json:"user"`
	}
	r, e := c.raw(ctx, "POST", "/auth/login", map[string]any{"email": email, "password": password, "device_id": c.session.Device}, &pre)
	if e != nil {
		return nil, e
	}
	var v struct {
		Access       string         `json:"access_token"`
		RefreshToken string         `json:"refresh_token"`
		Expires      int            `json:"expires_in"`
		ActiveScope  map[string]any `json:"active_scope"`
		User         map[string]any `json:"user"`
	}
	if pre.PreAuthToken != "" {
		if len(pre.AccessibleScopes) == 0 {
			return nil, errors.New("Akun tidak memiliki akses role atau arena")
		}
		chosen := pre.AccessibleScopes[0].ScopeID
		for _, s := range pre.AccessibleScopes {
			if s.ArenaID != nil && *s.ArenaID != "" {
				chosen = s.ScopeID
				break
			}
		}
		r, e = c.raw(ctx, "POST", "/auth/select-scope", map[string]any{
			"pre_auth_token": pre.PreAuthToken,
			"scope_id":       chosen,
			"device_id":      c.session.Device,
		}, &v)
		if e != nil {
			return nil, e
		}
	} else {
		v.Access = pre.Access
		v.RefreshToken = pre.RefreshToken
		v.Expires = pre.Expires
		v.User = pre.User
	}
	if e = c.tokens(r, struct {
		Access       string         `json:"access_token"`
		RefreshToken string         `json:"refresh_token"`
		Expires      int            `json:"expires_in"`
		User         map[string]any `json:"user"`
	}{v.Access, v.RefreshToken, v.Expires, v.User}); e != nil {
		return nil, e
	}
	if v.ActiveScope != nil && v.User != nil {
		v.User["active_scope"] = v.ActiveScope
	}
	return v.User, nil
}

func (c *Client) Call(ctx context.Context, method, path string, body, out any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session.Access == "" || time.Now().After(c.session.Expires.Add(-30*time.Second)) {
		if e := c.refresh(ctx); e != nil {
			return e
		}
	}
	_, e := c.raw(ctx, method, path, body, out)
	var f *Failure
	if errors.As(e, &f) && f.Status == http.StatusUnauthorized {
		if e = c.refresh(ctx); e != nil {
			return e
		}
		_, e = c.raw(ctx, method, path, body, out)
		if errors.As(e, &f) && f.Status == http.StatusUnauthorized {
			return &Failure{http.StatusUnauthorized, signInRequired}
		}
	}
	return e
}

func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, e := c.raw(ctx, "POST", "/auth/logout", map[string]any{}, nil)
	device := c.session.Device
	c.session = Session{Base: c.Base, Device: device}
	se := c.store.Save("cloud-session.bin", c.session, true)
	if e != nil {
		return fmt.Errorf("Sesi desktop dihapus; pencabutan cloud belum terkonfirmasi")
	}
	return se
}
