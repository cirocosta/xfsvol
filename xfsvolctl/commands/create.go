package commands

import (
	"gopkg.in/urfave/cli.v1"
)

var Create = cli.Command{
	Name:  "create",
	Usage: "Creates a XFSVOL managed volume",
}
