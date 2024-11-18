package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/goplugin/helmenv/chaos/experiments"
	"github.com/goplugin/helmenv/environment"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	e, err := environment.DeployOrLoadEnvironment(
		environment.NewPluginConfig(nil, "helmenv-programmatic-example", environment.DefaultGeth),
	)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	defer e.DeferTeardown()

	time.Sleep(10 * time.Second)
	_, err = e.ApplyChaosExperiment(&experiments.PodFailure{
		LabelKey:   "app",
		LabelValue: "plugin-node",
		Duration:   10 * time.Second,
	})
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	time.Sleep(10 * time.Second)
	if err := e.Chaos.StopAll(); err != nil {
		log.Error().Msg(err.Error())
		return
	}
}
