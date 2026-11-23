package vlc

import (
	"fmt"
	"unsafe"
)

// SetZoom changes the crop on the existing video output. The polling loop
// retries if the video dimensions were not available when the slider moved.
func (w *Worker) SetZoom(percent int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.zoom = percent
	w.applyZoom()
}

func (w *Worker) applyZoom() {
	if w.player == 0 || w.zoom == 0 || w.appliedZoom == w.zoom {
		return
	}
	if w.zoom == 100 {
		if w.appliedZoom != 0 {
			w.lib.text("libvlc_video_set_crop_geometry", w.player, "")
		}
		w.appliedZoom = w.zoom
		return
	}
	var width, height uint32
	if w.lib.call("libvlc_video_get_size", w.player, 0, uintptr(unsafe.Pointer(&width)), uintptr(unsafe.Pointer(&height))) != 0 || width < 2 || height < 2 {
		return
	}
	w.lib.text("libvlc_video_set_crop_geometry", w.player, centeredCrop(width, height, w.zoom))
	w.appliedZoom = w.zoom
}

func centeredCrop(width, height uint32, percent int) string {
	// Keep the crop even for subsampled video. VLC scales it to fit the pane.
	cropWidth := max(2, int(width)*50/percent*2)
	cropHeight := max(2, int(height)*50/percent*2)
	return fmt.Sprintf("%dx%d+%d+%d", cropWidth, cropHeight, (int(width)-cropWidth)/2, (int(height)-cropHeight)/2)
}
