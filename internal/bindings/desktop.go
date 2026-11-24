package bindings

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"

	"kicau.desktop/cctv/internal/container"
	"kicau.desktop/cctv/internal/model"
)

type Desktop struct {
	container *container.Container
	app       *application.App
	window    *application.WebviewWindow
	ctx       context.Context
	cancel    context.CancelFunc
}

func New() (*Desktop, error) {
	c, err := container.New()
	if err != nil {
		return nil, err
	}
	return NewWithContainer(c), nil
}

func NewWithContainer(c *container.Container) *Desktop {
	return &Desktop{
		container: c,
		ctx:       c.Ctx,
		cancel:    c.Cancel,
	}
}

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
