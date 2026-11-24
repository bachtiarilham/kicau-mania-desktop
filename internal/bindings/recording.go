package bindings

import (
	"errors"
	"strings"

	"kicau.desktop/cctv/internal/player"
)

func (d *Desktop) ExportRecording(id string, suggestedFilename string) (string, error) {
	if !player.ValidID(id) {
		return "", errors.New("ID rekaman tidak valid")
	}
	ext, label := ".mkv", "Video MKV"
	d.HidePlayers()

	filename := id + ext
	if suggestedFilename != "" {
		filename = suggestedFilename
		if !strings.HasSuffix(strings.ToLower(filename), ext) {
			filename += ext
		}
	}

	dlg := d.app.Dialog.SaveFile().SetFilename(filename).AddFilter(label, "*"+ext).AttachToWindow(d.window)
	cur := d.container.ConfigUC.GetRecordingsDir()
	if cur != "" {
		dlg.SetDirectory(cur)
	}

	dest, err := dlg.PromptForSingleSelection()
	if err != nil || dest == "" {
		return "", err
	}

	return d.container.RecordingUC.ExportRecording(d.ctx, id, filename, dest)
}

func (d *Desktop) SaveSessionRecordings(sessionID, sessionName, targetDir, format string) ([]string, error) {
	return d.container.RecordingUC.SaveSessionRecordings(d.ctx, sessionID, sessionName, targetDir, format)
}
