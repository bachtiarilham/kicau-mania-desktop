package bindings

import configModel "kicau.desktop/cctv/internal/config/model"

func (d *Desktop) GetConfig() configModel.Config {
	return d.container.ConfigUC.GetConfig()
}

func (d *Desktop) SaveConfig(cfg configModel.Config) error {
	if err := d.container.ConfigUC.SaveConfig(cfg); err != nil {
		return err
	}
	if d.container.Cloud != nil {
		d.container.Cloud.UseBase(cfg.CloudURL)
	}
	return nil
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
