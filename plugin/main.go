package main

import (
	"os"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog"

	v "github.com/docker/go-plugins-helpers/volume"
)

const (
	socketAddress = "/run/docker/plugins/xfsvol.sock"
)

type config struct {
	HostMountpoint string `arg:"--host-mountpoint,env:HOST_MOUNTPOINT,help:xfs-mounted filesystem to create volumes"`
	DefaultSize    string `arg:"--default-size,env:DEFAULT_SIZE,help:default size to use as quota"`
	Debug          bool   `arg:"env:DEBUG,help:enable debug logs"`
}

var (
	version string = "master-dev"
	logger         = zerolog.New(os.Stdout)
	args           = &config{
		HostMountpoint: "/mnt/xfs/volumes",
		DefaultSize:    "512M",
		Debug:          false,
	}
)

func main() {
	arg.MustParse(args)

	logger.Info().
		Str("version", version).
		Str("socket-address", socketAddress).
		Interface("args", args).
		Msg("initializing plugin")

	if args.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	d, err := NewDriver(DriverConfig{
		HostMountpoint: args.HostMountpoint,
		DefaultSize:    args.DefaultSize,
	})
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
