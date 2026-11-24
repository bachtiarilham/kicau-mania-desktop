package bindings

import "kicau.desktop/cctv/internal/model"

func (d *Desktop) GetConfig() model.Config {
	return d.container.ConfigUC.GetConfig()
}

func (d *Desktop) SaveConfig(cfg model.Config) error {
	return d.container.ConfigUC.SaveConfig(cfg)
}

func (d *Desktop) Preferences() string {
	return d.container.ConfigUC.Preferences()
}

func (d *Desktop) OpenRecordingsFolder() error {
	return d.container.ConfigUC.OpenRecordingsFolder()
}

func (d *Desktop) SelectRecordingsFolder() (string, error) {
	d.HidePlayers()
	dlg := d.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("Pilih Folder Penyimpanan Rekaman Video").
		AttachToWindow(d.window)

	cur := d.container.ConfigUC.GetRecordingsDir()
	if cur != "" {
		dlg.SetDirectory(cur)
	}
	dir, err := dlg.PromptForSingleSelection()
	if err != nil || dir == "" {
		return cur, err
	}
	_ = d.container.ConfigUC.SetRecordingsDir(dir)
	return dir, nil
}
