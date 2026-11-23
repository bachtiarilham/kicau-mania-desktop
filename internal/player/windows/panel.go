package windows

import (
	"errors"
	"golang.org/x/sys/windows"
	"kicau.desktop/cctv/internal/model"
	"math"
	"unsafe"
)

var user = windows.NewLazySystemDLL("user32.dll")
var create = user.NewProc("CreateWindowExW")
var position = user.NewProc("SetWindowPos")
var destroy = user.NewProc("DestroyWindow")
var show = user.NewProc("ShowWindow")
var dpi = user.NewProc("GetDpiForWindow")
var region = user.NewProc("SetWindowRgn")
var gdi = windows.NewLazySystemDLL("gdi32.dll")
var rectRegion = gdi.NewProc("CreateRectRgn")
var deleteObject = gdi.NewProc("DeleteObject")

func New(parent uintptr) (uintptr, error) {
	class, _ := windows.UTF16PtrFromString("STATIC")
	name, _ := windows.UTF16PtrFromString("")
	h, _, _ := create.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(name)), 0x40000000|0x02000000|0x04000000|0x00000004, 0, 0, 1, 1, parent, 0, 0, 0)
	if h == 0 {
		return 0, errors.New("Panel video Windows gagal dibuat")
	}
	return h, nil
}
func Bounds(parent, h uintptr, r model.Rect) {
	SetBounds(parent, h, r, true)
}
func SetBounds(parent, h uintptr, r model.Rect, visible bool) {
	if !visible || !r.Visible || r.Width <= 0 || r.Height <= 0 || r.ClipWidth <= 0 || r.ClipHeight <= 0 {
		show.Call(h, 0)
		return
	}
	d, _, _ := dpi.Call(parent)
	if d == 0 {
		d = 96
	}
	scale := float64(d) / 96
	px := func(v int) uintptr { return uintptr(int32(math.Round(float64(v) * scale))) }
	clip, _, _ := rectRegion.Call(px(r.ClipX), px(r.ClipY), px(r.ClipX+r.ClipWidth), px(r.ClipY+r.ClipHeight))
	ok, _, _ := region.Call(h, clip, 1)
	if ok == 0 {
		deleteObject.Call(clip)
	}
	position.Call(h, 0, px(r.X), px(r.Y), px(r.Width), px(r.Height), 0x0010|0x0040)
}
func Destroy(h uintptr) { destroy.Call(h) }
func Hide(h uintptr)    { show.Call(h, 0) }
