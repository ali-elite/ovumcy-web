package api

func (handler *Handler) oidcEnabled() bool {
	return handler != nil && handler.oidcService != nil && handler.oidcService.Enabled()
}

func (handler *Handler) localPublicAuthEnabled() bool {
	if handler == nil || handler.oidcService == nil {
		return true
	}
	return handler.oidcService.LocalPublicAuthEnabled()
}
