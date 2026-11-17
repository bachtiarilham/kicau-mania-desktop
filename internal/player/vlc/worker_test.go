package vlc

import (
	"golang.org/x/sys/windows"
	"kicau.desktop/cctv/internal/model"
	native "kicau.desktop/cctv/internal/player/windows"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// Opt-in native acceptance against an real HEVC NVR source. The NVR is read only; the operational engine is never started or stopped by this test.
func TestNativeRTSPEmbeddingAndCleanup(t *testing.T) {
	binary, source := os.Getenv("KICAU_TEST_VLC"), os.Getenv("KICAU_TEST_NVR_RTSP_URL")
	if binary == "" || source == "" {
		t.Skip("Set KICAU_TEST_VLC and isolated KICAU_TEST_NVR_RTSP_URL")
	}
	testNativePlayback(t, binary, source, 3*time.Second)
}

// Multiple real streams must advance concurrently, not just open successfully.
func TestNativeConcurrentRTSP(t *testing.T) {
	binary, sources := os.Getenv("KICAU_TEST_VLC"), os.Getenv("KICAU_TEST_NVR_RTSP_URLS")
	if binary == "" || sources == "" {
		t.Skip("Set KICAU_TEST_VLC and KICAU_TEST_NVR_RTSP_URLS (comma separated)")
	}
	for i, source := range strings.Split(sources, ",") {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()
			testNativePlayback(t, binary, strings.TrimSpace(source), 15*time.Second)
		})
	}
}

func testNativePlayback(t *testing.T, binary, source string, sample time.Duration) {
	t.Helper()
	user := windows.NewLazySystemDLL("user32.dll")
	type msg struct {
		Hwnd    uintptr
		Message uint32
		Wparam  uintptr
		Lparam  uintptr
		Time    uint32
		X, Y    int32
		Private uint32
	}
	ready := make(chan [2]uintptr, 1)
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(stopped)
		name, _ := windows.UTF16PtrFromString("STATIC")
		parent, _, _ := user.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(name)), 0x00cf0000, 0, 0, 800, 600, 0, 0, 0, 0)
		panel, e := native.New(parent)
		if e != nil {
			ready <- [2]uintptr{}
			return
		}
		native.Bounds(parent, panel, model.Rect{X: 10, Y: 10, Width: 640, Height: 360, ClipWidth: 640, ClipHeight: 360, Visible: true})
		ready <- [2]uintptr{parent, panel}
		defer native.Destroy(panel)
		defer native.Destroy(parent)
		for {
			select {
			case <-done:
				return
			default:
				var m msg
				for {
					ok, _, _ := user.NewProc("PeekMessageW").Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, 1)
					if ok == 0 {
						break
					}
					user.NewProc("TranslateMessage").Call(uintptr(unsafe.Pointer(&m)))
					user.NewProc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&m)))
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()
	handles := <-ready
	if handles[1] == 0 {
		t.Fatal("native panel creation failed")
	}
	defer func() { close(done); <-stopped }()
	parent, _, _ := user.NewProc("GetParent").Call(handles[1])
	if parent != handles[0] {
		t.Fatal("player panel not embedded in parent")
	}
	playing := make(chan struct{}, 1)
	cache := 50
	if raw := os.Getenv("KICAU_TEST_CACHE_MS"); raw != "" {
		cache, _ = strconv.Atoi(raw)
	}
	started := time.Now()
	worker, e := StartWithCache(binary, source, handles[1], "", cache, func(progress Progress) {
		if progress.State == "PLAYING" {
			select {
			case playing <- struct{}{}:
			default:
			}
		}
	})
	if e != nil {
		t.Fatal(e)
	}
	defer worker.Close()
	select {
	case <-playing:
		t.Logf("first advancing frame after %s (cache=%d ms)", time.Since(started).Round(time.Millisecond), cache)
	case <-time.After(25 * time.Second):
		t.Fatal("RTSP video did not reach VLC PLAYING")
	}
	if worker.lib.call("libvlc_media_player_get_hwnd", worker.player) != handles[1] {
		t.Fatal("VLC HWND does not match native panel")
	}
	video := false
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if worker.lib.call("libvlc_media_player_has_vout", worker.player) > 0 {
			video = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !video {
		t.Fatal("VLC reached playback without a video output")
	}
	before, ok := worker.Stats()
	if !ok {
		t.Fatal("native frame statistics unavailable")
	}
	time.Sleep(sample)
	after, ok := worker.Stats()
	if !ok || after.DecodedVideo <= before.DecodedVideo || after.DisplayedPictures <= before.DisplayedPictures {
		t.Fatalf("video failed to advance: before=%+v after=%+v", before, after)
	}
	t.Logf("%s native sample: decoded +%d, displayed +%d, lost +%d", sample, after.DecodedVideo-before.DecodedVideo, after.DisplayedPictures-before.DisplayedPictures, after.LostPictures-before.LostPictures)
	worker.Close()
	select {
	case <-worker.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("VLC worker survived shutdown")
	}
}
