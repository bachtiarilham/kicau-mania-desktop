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

func normalizeBase(base string) string {
	base = strings.TrimRight(base, "/")
	if !strings.HasSuffix(base, "/api/v1") {
		base += "/api/v1"
	}
	return base
}

func New(base string, store config.ConfigStore) *Client {
	base = normalizeBase(base)
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

// UseBase switches the running client and prevents reuse of credentials on another host.
func (c *Client) UseBase(base string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	base = normalizeBase(base)
	if base == c.Base {
		return
	}
	u, _ := url.Parse(base)
	c.Base, c.Origin = base, u.Scheme+"://"+u.Host
	c.session = Session{Base: base, Device: c.session.Device}
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

type tokenResponse struct {
	Access       string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Expires      int            `json:"expires_in"`
	ActiveScope  map[string]any `json:"active_scope"`
}

func (c *Client) tokenSession(res *http.Response, v tokenResponse) (Session, error) {
	if v.Access == "" || v.Expires < 1 {
		return Session{}, errors.New("Token cloud tidak valid")
	}
	next := Session{Base: c.Base, Device: c.session.Device, Access: v.Access,
		Expires: time.Now().Add(time.Duration(v.Expires) * time.Second)}
	for _, cookie := range res.Cookies() {
		if (cookie.Name == "__Host-kicau_refresh" || cookie.Name == "kicau_refresh") && cookie.Value != "" {
			next.Refresh, next.Cookie = cookie.Value, cookie.Name
			break
		}
	}
	if next.Refresh == "" && v.RefreshToken != "" {
		next.Refresh, next.Cookie = v.RefreshToken, "kicau_refresh"
		if c.session.Cookie != "" {
			next.Cookie = c.session.Cookie
		}
	}
	if next.Refresh == "" {
		return Session{}, errors.New("Cloud tidak mengirim refresh cookie")
	}
	return next, nil
}

func (c *Client) refresh(ctx context.Context) error {
	if c.session.Refresh == "" {
		return &Failure{http.StatusUnauthorized, signInRequired}
	}
	var v tokenResponse
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
	next, err := c.tokenSession(r, v)
	if err != nil {
		return err
	}
	c.session = next
	return c.store.Save("cloud-session.bin", next, true)
}

func (c *Client) Login(ctx context.Context, email, password string) (map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var pre struct {
		tokenResponse
		PreAuthToken     string `json:"pre_auth_token"`
		AccessibleScopes []struct {
			ScopeID   string  `json:"scope_id"`
			Role      string  `json:"role"`
			ScopeType string  `json:"scope_type"`
			ArenaID   *string `json:"arena_id"`
		} `json:"accessible_scopes"`
	}
	r, e := c.raw(ctx, "POST", "/auth/login", map[string]any{"email": email, "password": password, "device_id": c.session.Device}, &pre)
	if e != nil {
		return nil, e
	}
	v := pre.tokenResponse
	if pre.PreAuthToken != "" {
		chosen := ""
		for _, s := range pre.AccessibleScopes {
			role := strings.ToUpper(s.Role)
			if s.ScopeID == "" || (role != "ADMIN" && role != "OPERATOR" && role != "JUDGE" && role != "JURI") {
				continue
			}
			if chosen == "" {
				chosen = s.ScopeID
			}
			if s.ArenaID != nil && *s.ArenaID != "" {
				chosen = s.ScopeID
				break
			}
		}
		if chosen == "" {
			return nil, errors.New("Akun tidak memiliki akses Admin, Operator, atau Juri")
		}
		v = tokenResponse{}
		r, e = c.raw(ctx, "POST", "/auth/select-scope", map[string]any{
			"pre_auth_token": pre.PreAuthToken,
			"scope_id":       chosen,
			"device_id":      c.session.Device,
		}, &v)
		if e != nil {
			return nil, e
		}
	}
	next, e := c.tokenSession(r, v)
	if e != nil {
		return nil, e
	}
	previous := c.session
	c.session = next
	var profile map[string]any
	if _, e = c.raw(ctx, "GET", "/auth/me", nil, &profile); e != nil {
		c.session = previous
		var f *Failure
		if errors.As(e, &f) && f.Status == http.StatusUnauthorized {
			return nil, errors.New("Cloud menerima login, tetapi menolak token pada pemeriksaan sesi. Periksa deployment cloud-app.")
		}
		return nil, fmt.Errorf("Gagal memverifikasi sesi cloud: %w", e)
	}
	if id, _ := profile["id"].(string); id == "" {
		c.session = previous
		return nil, errors.New("Cloud tidak mengembalikan identitas pengguna")
	}
	if e = c.store.Save("cloud-session.bin", next, true); e != nil {
		c.session = previous
		return nil, e
	}
	if v.ActiveScope != nil {
		profile["active_scope"] = v.ActiveScope
	}
	return profile, nil
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
