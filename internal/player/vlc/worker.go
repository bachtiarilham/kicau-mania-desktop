// Package vlc embeds LibVLC 3.x directly in a native Windows child HWND.
package vlc

import (
	"errors"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type library struct {
	dll   *windows.DLL
	procs map[string]*windows.Proc
}

var libraries struct {
	sync.Mutex
	byPath map[string]*library
}

// Libraries stay loaded until process exit; VLC plugin threads must never run
// against an unloaded DLL. Instances and media players are released per pane.
func load(path string) (*library, error) {
	libraries.Lock()
	defer libraries.Unlock()
	if libraries.byPath == nil {
		libraries.byPath = map[string]*library{}
	}
	path = filepath.Clean(path)
	if l := libraries.byPath[path]; l != nil {
		return l, nil
	}
	if !filepath.IsAbs(path) {
		return nil, errors.New("Lokasi libvlc.dll harus absolut")
	}
	h, e := windows.LoadLibraryEx(path, 0, windows.LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR|windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
	if e != nil {
		return nil, errors.New("libvlc.dll gagal dimuat; periksa paket VLC amd64 dan folder plugins")
	}
	l := &library{dll: &windows.DLL{Name: path, Handle: h}, procs: map[string]*windows.Proc{}}
	for _, name := range []string{"libvlc_new", "libvlc_release", "libvlc_media_new_location", "libvlc_media_add_option", "libvlc_media_release", "libvlc_media_get_stats", "libvlc_media_player_new_from_media", "libvlc_media_player_set_hwnd", "libvlc_media_player_play", "libvlc_media_player_stop", "libvlc_media_player_release", "libvlc_media_player_get_state", "libvlc_media_player_has_vout", "libvlc_media_player_get_hwnd", "libvlc_audio_set_mute", "libvlc_video_set_mouse_input", "libvlc_video_set_key_input"} {
		p, e := l.dll.FindProc(name)
		if e != nil {
			return nil, errors.New("Versi libVLC tidak kompatibel; gunakan paket VLC 3.x")
		}
		l.procs[name] = p
	}
	libraries.byPath[path] = l
	return l, nil
}

//go:uintptrescapes
func (l *library) call(name string, args ...uintptr) uintptr {
	r, _, _ := l.procs[name].Call(args...)
	return r
}
func (l *library) text(name string, first uintptr, value string) uintptr {
	b, e := windows.BytePtrFromString(value)
	if e != nil {
		return 0
	}
	r := l.call(name, first, uintptr(unsafe.Pointer(b)))
	runtime.KeepAlive(b)
	return r
}

type Worker struct {
	mu                      sync.Mutex
	lib                     *library
	instance, player, media uintptr
	stop, done              chan struct{}
	once                    sync.Once
	proxyClose              func()
	muted                   bool
}

func (w *Worker) Done() <-chan struct{} { return w.done }

func (w *Worker) Stats() (mediaStats, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var value mediaStats
	if w.media == 0 {
		return value, false
	}
	ok := w.lib.call("libvlc_media_get_stats", w.media, uintptr(unsafe.Pointer(&value))) != 0
	return value, ok
}

var visibilityAPI = windows.NewLazySystemDLL("user32.dll")
var isWindowVisible = visibilityAPI.NewProc("IsWindowVisible")
var getAncestor = visibilityAPI.NewProc("GetAncestor")
var isIconic = visibilityAPI.NewProc("IsIconic")

func panelVisible(hwnd uintptr) bool {
	visible, _, _ := isWindowVisible.Call(hwnd)
	root, _, _ := getAncestor.Call(hwnd, 2)
	minimized, _, _ := isIconic.Call(root)
	return visible != 0 && minimized == 0
}

type Progress struct {
	State                    string
	FPS                      float64
	Decoded, Displayed, Lost int32
}

func Start(binary, source string, hwnd uintptr, headers string, event func(Progress)) (*Worker, error) {
	return StartWithCache(binary, source, hwnd, headers, 50, event)
}

// Cache applies only to live RTSP. Recorded-file playback keeps VLC defaults.
// Zero removes the requested input delay; source/decoder latency still exists.
func StartWithCache(binary, source string, hwnd uintptr, headers string, cacheMS int, event func(Progress)) (*Worker, error) {
	if cacheMS < 0 || cacheMS > 3000 {
		return nil, errors.New("Buffer live harus antara 0 dan 3000 ms")
	}
	l, e := load(binary)
	if e != nil {
		return nil, e
	}
	u, e := url.Parse(source)
	if e != nil || strings.ContainsAny(source, "\x00\r\n") {
		return nil, errors.New("Sumber player tidak valid")
	}
	cleanup := func() {}
	if headers != "" {
		source, cleanup, e = recordingProxy(source, headers)
		if e != nil {
			return nil, e
		}
	}
	args := []string{"--ignore-config", "--no-video-title-show", "--no-osd", "--no-snapshot-preview", "--quiet"}
	live := u.Scheme == "rtsp" || u.Scheme == "rtsps"
	if live {
		// Use the Windows DXVA2/D3D9 path for the embedded multi-camera grid.
		// Keep decoding on the GPU without sharing D3D11 surfaces with WebView2.
		// This only changes local presentation; no media is encoded or converted.
		args = append(args, "--vout=direct3d9", "--avcodec-hw=dxva2", "--avcodec-threads=2")
		args = append(args, "--rtsp-tcp", "--network-caching="+strconv.Itoa(cacheMS), "--live-caching="+strconv.Itoa(cacheMS), "--clock-jitter=0", "--drop-late-frames")
	}
	storage := make([]*byte, len(args))
	argv := make([]uintptr, len(args))
	for i, v := range args {
		storage[i], _ = windows.BytePtrFromString(v)
		argv[i] = uintptr(unsafe.Pointer(storage[i]))
	}
	instance := l.call("libvlc_new", uintptr(len(argv)), uintptr(unsafe.Pointer(&argv[0])))
	runtime.KeepAlive(storage)
	runtime.KeepAlive(argv)
	if instance == 0 {
		cleanup()
		return nil, errors.New("VLC gagal diinisialisasi")
	}
	media := l.text("libvlc_media_new_location", instance, source)
	if media == 0 {
		l.call("libvlc_release", instance)
		cleanup()
		return nil, errors.New("Sumber VLC tidak tersedia")
	}
	l.text("libvlc_media_add_option", media, ":no-audio-time-stretch")
	if u.Scheme == "rtsp" || u.Scheme == "rtsps" {
		l.text("libvlc_media_add_option", media, ":network-caching="+strconv.Itoa(cacheMS))
		l.text("libvlc_media_add_option", media, ":live-caching="+strconv.Itoa(cacheMS))
	}
	player := l.call("libvlc_media_player_new_from_media", media)
	if player == 0 {
		l.call("libvlc_media_release", media)
		l.call("libvlc_release", instance)
		cleanup()
		return nil, errors.New("Player VLC gagal dibuat")
	}
	l.call("libvlc_media_player_set_hwnd", player, hwnd)
	l.call("libvlc_video_set_mouse_input", player, 0)
	l.call("libvlc_video_set_key_input", player, 0)
	l.call("libvlc_audio_set_mute", player, 1)
	w := &Worker{lib: l, instance: instance, player: player, media: media, stop: make(chan struct{}), done: make(chan struct{}), proxyClose: cleanup, muted: true}
	if int32(l.call("libvlc_media_player_play", player)) != 0 {
		w.release()
		return nil, errors.New("VLC tidak dapat memulai playback")
	}
	go func() {
		defer close(w.done)
		tick := time.NewTicker(500 * time.Millisecond)
		defer tick.Stop()
		previous := ""
		lastReport := time.Now()
		lastDisplayed := int32(0)
		monitor := frameMonitor{started: time.Now(), progressed: time.Now()}
		for {
			select {
			case <-w.stop:
				return
			case <-tick.C:
				w.mu.Lock()
				state := l.call("libvlc_media_player_get_state", player)
				if state == 3 && previous != "PLAYING" {
					mute := uintptr(0)
					if w.muted {
						mute = 1
					}
					l.call("libvlc_audio_set_mute", player, mute)
				}
				var stats mediaStats
				statsOK := live && l.call("libvlc_media_get_stats", media, uintptr(unsafe.Pointer(&stats))) != 0
				w.mu.Unlock()
				status := "CONNECTING"
				reconnect := false
				switch state {
				case 2:
					status = "BUFFERING"
				case 3:
					status = "PLAYING"
				case 6:
					status = "ENDED"
				case 7:
					status = "FAILED"
				}
				if live && state != 6 && state != 7 {
					if statsOK {
						count := stats.DisplayedPictures
						if !panelVisible(hwnd) {
							count = stats.DecodedVideo
						}
						status, reconnect = monitor.observe(time.Now(), count)
					} else {
						status = "WAITING_VIDEO"
						if time.Since(monitor.started) >= 20*time.Second {
							status, reconnect = "RECONNECTING", true
						}
					}
				}
				now := time.Now()
				if status != previous || now.Sub(lastReport) >= 2*time.Second {
					previous = status
					fps := float64(stats.DisplayedPictures-lastDisplayed) / now.Sub(lastReport).Seconds()
					if fps < 0 {
						fps = 0
					}
					lastReport, lastDisplayed = now, stats.DisplayedPictures
					event(Progress{State: status, FPS: fps, Decoded: stats.DecodedVideo, Displayed: stats.DisplayedPictures, Lost: stats.LostPictures})
				}
				if reconnect || state == 6 || state == 7 {
					return
				}
			}
		}
	}()
	return w, nil
}
func (w *Worker) Command(command []any) error {
	if len(command) != 3 || command[0] != "set_property" || command[1] != "mute" {
		return errors.New("Perintah VLC tidak didukung")
	}
	muted, ok := command[2].(bool)
	if !ok {
		return errors.New("Status mute tidak valid")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.player == 0 {
		return errors.New("Player ditutup")
	}
	w.muted = muted
	v := uintptr(0)
	if muted {
		v = 1
	}
	w.lib.call("libvlc_audio_set_mute", w.player, v)
	return nil
}
func (w *Worker) release() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.player != 0 {
		w.lib.call("libvlc_media_player_stop", w.player)
		w.lib.call("libvlc_media_player_release", w.player)
		w.player = 0
	}
	if w.media != 0 {
		w.lib.call("libvlc_media_release", w.media)
		w.media = 0
	}
	if w.instance != 0 {
		w.lib.call("libvlc_release", w.instance)
		w.instance = 0
	}
	w.proxyClose()
}
func (w *Worker) Close() { w.once.Do(func() { close(w.stop); <-w.done; w.release() }) }
