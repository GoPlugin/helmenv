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
		environment.NewPluginConfig(nil, "helmenv-load-env", environment.DefaultGeth),
	)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	defer e.DeferTeardown()

	loadedEnv, err := environment.LoadEnvironment(e.Config)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	if err := loadedEnv.ConnectAll(); err != nil {
		log.Error().Msg(err.Error())
		return
	}
	remoteURLs, err := loadedEnv.Charts.Connections("geth").RemoteURLsByPort("http-rpc", environment.HTTP)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	log.Info().Interface("URLs", remoteURLs).Msg("Remote Geth")

	localURLs, err := loadedEnv.Charts.Connections("geth").LocalURLsByPort("http-rpc", environment.HTTP)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	log.Info().Interface("URLs", localURLs).Msg("Local Geth")
}
