package commands

import (
	"gopkg.in/urfave/cli.v1"
)

var Ls = cli.Command{
	Name:  "ls",
	Usage: "Lists the volumes managed by 'xfsvol' plugin",
}
