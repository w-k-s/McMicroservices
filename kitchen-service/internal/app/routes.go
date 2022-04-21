package app

func (a *App) configureRoutes() {
	a.router.HandleFunc("/health", a.HealthCheck())
	a.router.HandleFunc("/kitchen/api/v1/stock", a.GetStock())
}
