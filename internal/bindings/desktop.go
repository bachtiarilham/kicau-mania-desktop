package bindings

import (
	"context"
	"errors"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"kicau.desktop/cctv/internal/container"
	"kicau.desktop/cctv/internal/model"
	"kicau.desktop/cctv/internal/player"
)

// Desktop is the Wails RPC Delivery Facade that exposes methods to the frontend.
// It delegates all domain logic to the injected Clean Architecture usecases.
type Desktop struct {
	container *container.Container
	app       *application.App
	window    *application.WebviewWindow
	ctx       context.Context
	cancel    context.CancelFunc
}

// New creates a new Desktop delivery facade with fully initialized dependencies.
func New() (*Desktop, error) {
	c, err := container.New()
	if err != nil {
		return nil, err
	}
	return NewWithContainer(c), nil
}

// NewWithContainer creates a Desktop facade using a provided container (for testing and DI).
func NewWithContainer(c *container.Container) *Desktop {
	return &Desktop{
		container: c,
		ctx:       c.Ctx,
		cancel:    c.Cancel,
	}
}

// attach binds the native window handle and event emitter to the player driver.
func (d *Desktop) attach(a *application.App, w *application.WebviewWindow) {
	d.app = a
	d.window = w
	d.container.AttachPlayer(
		func() uintptr { return uintptr(w.NativeWindow()) },
		func(s model.PlayerStatus) { a.Event.Emit("player:status", s) },
	)
	go d.container.EngineUC.SubscribeMetrics(d.ctx, func(v any) {
		a.Event.Emit("engine:metric", v)
	})
}

// --- Config Domain ---

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

// --- Auth Domain ---

func (d *Desktop) Login(email, password string) (map[string]any, error) {
	return d.container.AuthUC.Login(d.ctx, email, password)
}

func (d *Desktop) Logout() error {
	return d.container.AuthUC.Logout(d.ctx)
}

func (d *Desktop) Cloud(path string) (any, error) {
	return d.container.AuthUC.Cloud(d.ctx, path)
}

func (d *Desktop) Revoke(engineID string) error {
	return d.container.AuthUC.Revoke(d.ctx, engineID)
}

// --- Engine Domain ---

func (d *Desktop) Bootstrap() (map[string]any, error) {
	return d.container.EngineUC.Bootstrap(d.ctx)
}

func (d *Desktop) Engine(method, path string, body map[string]any) (any, error) {
	return d.container.EngineUC.Engine(d.ctx, method, path, body)
}

func (d *Desktop) Pair(arena string) (any, error) {
	return d.container.EngineUC.Pair(d.ctx, arena)
}

func (d *Desktop) BindSession(in map[string]any) (any, error) {
	return d.container.EngineUC.BindSession(d.ctx, in)
}

func (d *Desktop) NewSessionID() string {
	return d.container.EngineUC.NewSessionID()
}

func (d *Desktop) LocalRecordings(cursor string) (any, error) {
	return d.container.EngineUC.LocalRecordings(d.ctx, cursor)
}

func (d *Desktop) IsActive() bool {
	return d.container.EngineUC.IsActive()
}

func (d *Desktop) SetActive(active bool) {
	d.container.EngineUC.SetActive(active)
}

// --- Media & Player Domain ---

func (d *Desktop) SyncPlayer(id string, r model.Rect) error {
	return d.SyncPlayerWithZoom(id, r, 100)
}

func (d *Desktop) SyncZoomPlayer(id string, r model.Rect) error {
	return d.SyncPlayerWithZoom(id, r, 200)
}

func (d *Desktop) SyncPlayerWithZoom(id string, r model.Rect, zoom int) error {
	if d.container.PlayerUC == nil {
		return errors.New("Player belum siap")
	}
	return d.container.PlayerUC.SyncPlayerWithZoom(d.ctx, id, r, zoom)
}

func (d *Desktop) StopPlayer(id string) {
	if d.container.PlayerUC != nil {
		d.container.PlayerUC.StopPlayer(id)
	}
}

func (d *Desktop) HidePlayers() {
	if d.container.PlayerUC != nil {
		d.container.PlayerUC.HidePlayers()
	}
}

func (d *Desktop) Mute(id string, muted bool) error {
	if d.container.PlayerUC == nil {
		return nil
	}
	return d.container.PlayerUC.Mute(id, muted)
}

func (d *Desktop) PlayRecording(id string, r model.Rect) error {
	if d.container.PlayerUC == nil {
		return errors.New("Player belum siap")
	}
	return d.container.PlayerUC.PlayRecording(id, r)
}

// --- Recording Domain ---

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
