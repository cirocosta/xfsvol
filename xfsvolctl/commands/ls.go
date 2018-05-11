package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cirocosta/xfsvol/manager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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

            xfsvolctl create \
                --root /mnt/xfs
                --name myvol
                --size 10M

            xfsvolctl ls \
                --root /mnt/xfs

            NAME      QUOTA
            myvol     10M
	`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "root, r",
			Usage: "Root of the volume listing",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Whether debug logs should be displayed",
		},
	},
	Action: lsAction,
}

func lsAction(c *cli.Context) (err error) {
	var (
		root  = c.String("root")
		debug = c.Bool("debug")
	)

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if root == "" {
		cli.ShowCommandHelp(c, "ls")
		err = cli.NewExitError("All parameters must be set", 1)
		return
	}

	mgr, err := manager.New(manager.Config{
		Root: root,
	})
	if err != nil {
		err = cli.NewExitError(errors.Wrapf(err,
			"Couldn't initiate manager"), 1)
		return
	}

	vols, err := mgr.List()
	if err != nil {
		err = cli.NewExitError(errors.Wrapf(err,
			"Couldn't list volumes under root %s", root), 1)
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
}
