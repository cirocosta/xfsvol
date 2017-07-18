package commands

import (
  "github.com/cirocosta/xfsvol/manager"
	"gopkg.in/urfave/cli.v1"
)

var Ls = cli.Command{
	Name:  "ls",
	Usage: "Lists the volumes managed by 'xfsvol' plugin",
  Action: func (c *cli.Context) (err error) {
    _, err = manager.New(manager.Config{
      Root: "/mnt/xfs",
    })
    if err != nil {
      return
    }

    return
  },
}
