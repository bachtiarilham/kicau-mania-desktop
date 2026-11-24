package windows

import (
	"runtime"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
	"kicau.desktop/cctv/internal/player/model"
)

func TestNativeRegionStaysInsidePane(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	class, _ := windows.UTF16PtrFromString("STATIC")
	parent, _, err := create.Call(0, uintptr(unsafe.Pointer(class)), 0, 0x80000000, 0, 0, 400, 300, 0, 0, 0, 0)
	if parent == 0 {
		t.Fatal(err)
	}
	defer Destroy(parent)
	pane, err := New(parent)
	if err != nil {
		t.Fatal(err)
	}
	defer Destroy(pane)
	Bounds(parent, pane, model.Rect{
		Width: 200, Height: 100, ClipX: -10, ClipY: 7, ClipWidth: 230, ClipHeight: 120,
		Visible: true, DevicePixels: true,
	})
	var got struct{ Left, Top, Right, Bottom int32 }
	kind, _, err := user.NewProc("GetWindowRgnBox").Call(pane, uintptr(unsafe.Pointer(&got)))
	if kind == 0 {
		t.Fatal(err)
	}
	if got.Left != 0 || got.Top != 7 || got.Right != 200 || got.Bottom != 100 {
		t.Fatalf("unexpected window region: %+v", got)
	}
}
