package bindings

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"golang.org/x/sys/windows"
	"unsafe"
)

// Wire deliberately lives outside the bound service's method set.
func Wire(d *Desktop, a *application.App, w *application.WebviewWindow) {
	d.attach(a, w)
	w.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if d.IsActive() {
			text, _ := windows.UTF16PtrFromString("Tutup desktop? Rekaman dan upload tetap berjalan di local-app.")
			title, _ := windows.UTF16PtrFromString("Sesi media masih aktif")
			proc := windows.NewLazySystemDLL("user32.dll").NewProc("MessageBoxW")
			answer, _, _ := proc.Call(uintptr(w.NativeWindow()), uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x00000004|0x00000030|0x00000100)
			if answer != 6 {
				event.Cancel()
			}
		}
	})
	a.OnShutdown(func() {
		d.cancel()
		if d.container != nil && d.container.PlayerUC != nil {
			d.container.PlayerUC.Shutdown()
		}
	})
}
