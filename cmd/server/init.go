package main

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/app"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
)

func initializeApp() (*app.Server, error) {
	server, err := app.NewContainer().
		LoadConfig().
		InitLogger().
		InitRepository().
		InitStore().
		InitService().
		InitHandler().
		RestoreData().
		MustBootstrap().
		Build()

	if err != nil {
		return nil, err
	}

	defer zl.Log.Sync()

	return server, nil
}
