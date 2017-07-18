package commands

import (
	"gopkg.in/urfave/cli.v1"
)

var Delete = cli.Command{
	Name:  "delete",
	Usage: "Deletes a volume managed by 'xfsvol' plugin",
}
