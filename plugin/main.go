package main

import (
	"os"

	"github.com/rs/zerolog"

	v "github.com/docker/go-plugins-helpers/volume"
)

const (
	socketAddress = "/run/docker/plugins/xfsvol.sock"
)

var (
	version string = "master-dev"
	logger         = zerolog.New(os.Stdout)
)

func main() {
	logger.Info().
		Str("version", version).
		Str("socket-address", socketAddress).
		Msg("initializing plugin")

	d, err := newNfsVolDriver()
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to initialize XFS volume driver")
		os.Exit(1)
	}

	h := v.NewHandler(d)
	err = h.ServeUnix(socketAddress, 0)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("socket-address", socketAddress).
			Msg("failed to server volume plugin api over unix socket")
		os.Exit(1)
	}

	return
}
