package bindings

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
