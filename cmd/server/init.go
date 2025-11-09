package main

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/app"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
)

func initializeApp() (*app.Server, error) {
	container, err := app.NewContainerWithOptions(
		app.WithConfig(),
		app.WithLogger(),
		app.WithRepository(),
		app.WithStore(),
		app.WithService(),
		app.WithAuditService(),
		app.WithCache(),
		app.WithHandler(),
		app.WithRestoreData(),
		app.WithBootstrap())
	if err != nil {
		return nil, err
	}
	server := app.Build(container)
	defer zl.Log.Sync()

	return server, nil
}
