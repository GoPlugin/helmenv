package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/goplugin/helmenv/environment"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	e, err := environment.DeployOrLoadEnvironment(
		environment.NewPluginConfig(nil, "helmenv-dump-example", environment.DefaultGeth),
	)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	defer e.DeferTeardown()

	if err := e.Artifacts.DumpTestResult("test_1", "plugin"); err != nil {
		log.Error().Msg(err.Error())
	}
}
