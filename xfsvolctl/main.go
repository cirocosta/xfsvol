package main

import (
	"os"

	"github.com/cirocosta/xfsvol/xfsvolctl/commands"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "xfsvolctl"
	app.Usage = "Controls the 'xfsvol' volume plugin"
	app.Commands = []cli.Command{
		commands.Ls,
	}
	app.Run(os.Args)
}
