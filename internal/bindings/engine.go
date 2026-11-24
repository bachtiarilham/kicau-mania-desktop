package bindings

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
