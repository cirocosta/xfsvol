package main

import (
	"os"

	"github.com/cirocosta/xfsvol/xfsvolctl/commands"
	"github.com/rs/zerolog"
	"gopkg.in/urfave/cli.v1"
)

var (
	version string = "master-dev"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	app := cli.NewApp()
	app.Name = "xfsvolctl"
	app.Version = version
	app.Usage = "Controls the 'xfsvol' volume plugin"
	app.Commands = []cli.Command{
		commands.Ls,
		commands.Create,
		commands.Delete,
	}
	app.Run(os.Args)
}
