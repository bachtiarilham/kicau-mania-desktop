package vlc

import "testing"

func TestCenteredCrop(t *testing.T) {
	for _, tc := range []struct {
		width, height uint32
		percent       int
		want          string
	}{
		{1920, 1080, 150, "1280x720+320+180"},
		{1920, 1080, 300, "640x360+640+360"},
		{641, 361, 200, "320x180+160+90"},
	} {
		if got := centeredCrop(tc.width, tc.height, tc.percent); got != tc.want {
			t.Errorf("crop at %d%% = %q, want %q", tc.percent, got, tc.want)
		}
	}
}
