package repository

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"kicau.desktop/cctv/internal/config/model"
)

type Store struct {
	mu  sync.Mutex
	Dir string
}

func New() (*Store, error) {
	dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "KicauManiaCCTV")
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	return &Store{Dir: dir}, nil
}

func protect(b []byte, encrypt bool) ([]byte, error) {
	if len(b) == 0 {
		return nil, errors.New("Secret kosong")
	}
	in := windows.DataBlob{Size: uint32(len(b)), Data: &b[0]}
	var out windows.DataBlob
	var e error
	if encrypt {
		e = windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	} else {
		e = windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	}
	if e != nil {
		return nil, errors.New("Proteksi credential Windows gagal")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}

func (s *Store) Save(name string, v any, secret bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	if secret {
		b, e = protect(b, true)
		if e != nil {
			return e
		}
	}
	f, e := os.CreateTemp(s.Dir, ".pending-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	return os.Rename(f.Name(), filepath.Join(s.Dir, name))
}

func (s *Store) Load(name string, v any, secret bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, e := os.ReadFile(filepath.Join(s.Dir, name))
	if e != nil {
		return e
	}
	if secret {
		b, e = protect(b, false)
		if e != nil {
			return e
		}
	}
	return json.Unmarshal(b, v)
}

func Default() model.Config {
	exe, _ := os.Executable()
	videosDir := filepath.Join(os.Getenv("USERPROFILE"), "Videos", "KicauMania")
	if os.Getenv("USERPROFILE") == "" {
		videosDir = filepath.Join(filepath.Dir(exe), "recordings")
	}
	return model.Config{
		LocalURL:      "http://127.0.0.1:8090",
		RTSPBase:      "rtsp://127.0.0.1:8554",
		CloudURL:      "https://dev.megaallstararena.com/api/v1",
		VLC:           filepath.Join(filepath.Dir(exe), "vlc", "libvlc.dll"),
		LiveCacheMS:   50,
		RecordingsDir: videosDir,
	}
}

func Validate(c model.Config) error {
	if c.LiveCacheMS < 0 || c.LiveCacheMS > 3000 {
		return errors.New("Buffer live harus antara 0 dan 3000 ms")
	}
	for _, item := range []struct{ raw, scheme string }{{c.LocalURL, "http"}, {c.RTSPBase, "rtsp"}} {
		u, e := url.Parse(item.raw)
		if e != nil || u.User != nil || u.Scheme != item.scheme || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return errors.New("Alamat engine harus berupa URL origin tanpa path")
		}
		ip := net.ParseIP(u.Hostname())
		if u.Hostname() != "localhost" && (ip == nil || (!ip.IsLoopback() && !ip.IsPrivate())) {
			return errors.New("Engine harus berjalan pada localhost atau jaringan lokal (LAN)")
		}
	}
	u, e := url.Parse(c.CloudURL)
	isLocal := u != nil && u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1")
	if e != nil || (u.Scheme != "https" && !isLocal) || u.User != nil || u.Hostname() == "" || u.RawQuery != "" || u.Fragment != "" || (strings.TrimRight(u.Path, "/") != "/api/v1" && strings.TrimRight(u.Path, "/") != "") {
		return errors.New("Cloud URL harus HTTPS origin (atau http://localhost untuk dev) dengan path /api/v1")
	}
	if !filepath.IsAbs(c.VLC) {
		return errors.New("Lokasi libvlc.dll harus path absolut")
	}
	if c.RecordingsDir != "" && !filepath.IsAbs(c.RecordingsDir) {
		return errors.New("Lokasi folder rekaman harus path absolut")
	}
	return nil
}
