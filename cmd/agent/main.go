package main

import (
	"fmt"
	"log"

	"github.com/bigsm0uk/metrics-alert-server/internal/app"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	app, err := InitializeApp()
	if err != nil {
		log.Fatal(err)
	}
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func InitializeApp() (*app.Agent, error) {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		return nil, err
	}
	logger := zl.NewLogger(cfg.Env)
	defer logger.Sync()

	return app.NewAgent(cfg, logger), nil
}

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}

	date := buildDate
	if date == "" {
		date = "N/A"
	}

	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
