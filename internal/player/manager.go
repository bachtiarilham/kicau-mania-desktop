package player

import (
	"errors"
	"github.com/wailsapp/wails/v3/pkg/application"
	"kicau.desktop/cctv/internal/model"
	"kicau.desktop/cctv/internal/player/vlc"
	native "kicau.desktop/cctv/internal/player/windows"
	"strings"
	"sync"
	"time"
)

type pane struct {
	handle    uintptr
	worker    *vlc.Worker
	closed    bool
	mu        sync.Mutex
	stop      chan struct{}
	bounds    model.Rect
	hasBounds bool
	ready     bool
}
type Manager struct {
	mu          sync.Mutex
	panes       map[string]*pane
	Parent      func() uintptr
	Binary      string
	LiveCacheMS int
	Emit        func(model.PlayerStatus)
	audible     string
}

func New(binary string, parent func() uintptr, emit func(model.PlayerStatus)) *Manager {
	return &Manager{panes: map[string]*pane{}, Parent: parent, Binary: binary, Emit: emit}
}
func (m *Manager) Sync(id, source, headers string, r model.Rect) error {
	m.mu.Lock()
	p := m.panes[id]
	if p == nil {
		if len(m.panes) >= 24 {
			m.mu.Unlock()
			return errors.New("Maksimal 24 player aktif")
		}
		p = &pane{stop: make(chan struct{})}
		m.panes[id] = p
	}
	m.mu.Unlock()
	var err error
	created := false
	application.InvokeSync(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.closed {
			return
		}
		parent := m.Parent()
		if parent == 0 {
			err = errors.New("Window belum siap")
			return
		}
		if p.handle == 0 {
			p.handle, err = native.New(parent)
			created = err == nil
		}
		if err == nil && (!p.hasBounds || p.bounds != r) {
			native.SetBounds(parent, p.handle, r, p.ready)
			p.bounds, p.hasBounds = r, true
		}
	})
	if err != nil {
		m.Stop(id)
		return err
	}
	if created {
		m.Emit(model.PlayerStatus{ChannelID: id, State: "CONNECTING"})
		go func() {
			for delay := time.Second; ; {
				started := time.Now()
				w, e := vlc.StartWithCache(m.Binary, source, p.handle, headers, m.LiveCacheMS, func(progress vlc.Progress) {
					isReady := progress.State == "PLAYING" && progress.Displayed > 0
					p.mu.Lock()
					if isReady != p.ready && !p.closed && p.handle != 0 {
						p.ready = isReady
						parent := m.Parent()
						application.InvokeSync(func() {
							native.SetBounds(parent, p.handle, p.bounds, isReady)
						})
					}
					p.mu.Unlock()

					state := progress.State
					message := ""
					switch state {
					case "STALLED":
						message = "Frame video tidak bertambah"
					case "RECONNECTING":
						message = "Menghubungkan ulang player RTSP"
					case "WAITING_VIDEO":
						message = "Menunggu frame video dari NVR"
					case "ENDED":
						message = "Menghubungkan kembali RTSP…"
					case "FAILED":
						message = "Kamera offline atau tidak terjangkau"
					}
					m.Emit(model.PlayerStatus{ChannelID: id, State: state, Message: message, FPS: progress.FPS, Decoded: progress.Decoded, Displayed: progress.Displayed, Lost: progress.Lost})
				})
				p.mu.Lock()
				if p.closed {
					if w != nil {
						w.Close()
					}
					p.mu.Unlock()
					return
				}
				p.worker = w
				p.mu.Unlock()
				if w != nil {
					m.mu.Lock()
					audible := m.audible == id
					m.mu.Unlock()
					_ = w.Command([]any{"set_property", "mute", !audible})
				}
				if e != nil {
					m.Emit(model.PlayerStatus{ChannelID: id, State: "FAILED", Message: e.Error()})
				}
				if w != nil {
					select {
					case <-p.stop:
						w.Close()
						return
					case <-w.Done():
					}
					w.Close()
				}
				p.mu.Lock()
				p.ready = false
				if p.handle != 0 && !p.closed {
					application.InvokeSync(func() { native.Hide(p.handle) })
				}
				p.mu.Unlock()
				if !strings.HasPrefix(source, "rtsp://") && !strings.HasPrefix(source, "rtsps://") {
					return
				}
				if time.Since(started) > 30*time.Second {
					delay = time.Second
				}
				select {
				case <-p.stop:
					return
				case <-time.After(delay):
				}
				if delay < 30*time.Second {
					delay = min(delay*2, 30*time.Second)
				}
			}
		}()
	}
	return nil
}
func (m *Manager) Stop(id string) {
	m.mu.Lock()
	p := m.panes[id]
	delete(m.panes, id)
	if m.audible == id {
		m.audible = ""
	}
	m.mu.Unlock()
	if p == nil {
		return
	}
	p.mu.Lock()
	if !p.closed {
		close(p.stop)
	}
	p.closed = true
	if p.worker != nil {
		p.worker.Close()
	}
	h := p.handle
	p.mu.Unlock()
	if h != 0 {
		application.InvokeSync(func() { native.Destroy(h) })
	}
}
func (m *Manager) Mute(id string, muted bool) error {
	m.mu.Lock()
	if muted {
		if m.audible == id {
			m.audible = ""
		}
	} else {
		m.audible = id
	}
	list := make(map[string]*pane, len(m.panes))
	for k, v := range m.panes {
		list[k] = v
	}
	m.mu.Unlock()
	for k, p := range list {
		p.mu.Lock()
		if p.worker != nil {
			_ = p.worker.Command([]any{"set_property", "mute", muted || k != id})
		}
		p.mu.Unlock()
	}
	return nil
}
func (m *Manager) Hide() {
	m.mu.Lock()
	list := make([]*pane, 0, len(m.panes))
	for _, p := range m.panes {
		list = append(list, p)
	}
	m.mu.Unlock()
	application.InvokeSync(func() {
		for _, p := range list {
			p.mu.Lock()
			if p.handle != 0 {
				native.Hide(p.handle)
				p.hasBounds = false
				p.ready = false
			}
			p.mu.Unlock()
		}
	})
}
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.panes {
		p.mu.Lock()
		if !p.closed {
			close(p.stop)
		}
		p.closed = true
		if p.worker != nil {
			p.worker.Close()
		}
		p.mu.Unlock()
	}
}
func ValidID(id string) bool {
	if len(id) < 1 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-", r) {
			return false
		}
	}
	return true
}
