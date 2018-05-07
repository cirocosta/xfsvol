package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cirocosta/xfsvol/manager"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

var Ls = cli.Command{
	Name:  "ls",
	Usage: "Lists the volumes managed by 'xfsvol' plugin",
	Description: `Lists the volumes created with XFS pquotas.
    Retrieve a list of the volumes created by 'xfsvol' Docker
    plugin or the 'xfsvolctl' command.

    Volumes are listed by their names relative to a root as
    well as the sizes assigned as project quota in XFS.

    Examples:

    1. create a volume with limit of 10M and then see it in
       the list of volumes:

       xfsvolctl create --root /mnt/xfs --name myvol --size 10M
       xfsvolctl ls --root /mnt/xfs

       NAME      QUOTA
       myvol     10M
	`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "root, r",
			Usage: "Root of the volume listing",
		},
	},
	Action: func(c *cli.Context) (err error) {
		var root = c.String("root")
		if root == "" {
			cli.ShowCommandHelp(c, "ls")
			err = errors.Errorf("All parameters must be set.")
			return
		}

		mgr, err := manager.New(manager.Config{
			Root: root,
		})
		if err != nil {
			err = errors.Wrapf(err,
				"Couldn't initiate manager")
			return
		}

		vols, err := mgr.List()
		if err != nil {
			err = errors.Wrapf(err,
				"Couldn't list volumes under root %s",
				root)
			return
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)
		fmt.Fprintln(w, "NAME\tBLK-QUOTA\tINODE-QUOTA\t")

		for _, vol := range vols {
			fmt.Fprintf(w, "%s\t%s\t%d\n",
				vol.Name,
				manager.HumanSize(vol.Size),
				vol.INode)
		}
		w.Flush()
		return
	},
}
