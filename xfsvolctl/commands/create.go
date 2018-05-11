package commands

import (
	"github.com/cirocosta/xfsvol/manager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gopkg.in/urfave/cli.v1"
)

var Create = cli.Command{
	Name:  "create",
	Usage: "Creates a volume with XFS project quota enforcement",
	Description: `Creates a docker volume with XFS pquotas.
   XFS project quota allows one to implement directory
   tree quota and accounting. This allows one to specify
   an intended size to limit the size of the content below
   a given tree.

   Examples:

     1. create a volume with an upper limit of 10M and no limits
        in number of inodes:

            lsblk
            NAME   MAJ:MIN RM   SIZE RO TYPE MOUNTPOINT
            loop0    7:0    0     2G  0 loop /mnt/xfs

            mount | grep xfs
            /dev/loop0 on /mnt/xfs type xfs (rw,relatime,attr2,inode64,prjquota)

            xfsvolctl create \
                --root /mnt/xfs \
                --name myvol \
                --size 10M

   Note:
     In order to have the creation functioning you must first have a
     mount point in the filesystem that is mounted on top of XFS and
     has 'pquota' set.
    `,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "Name of the volume to create",
		},
		cli.StringFlag{
			Name:  "size, s",
			Usage: "Size of the XFS project quota to apply (e.g.: 50M)",
		},
		cli.Uint64Flag{
			Name:  "inode, i",
			Usage: "Maximum number of INodes that can be created",
		},
		cli.StringFlag{
			Name:  "root, r",
			Usage: "Root of the volume creation (under an xfs filesystem)",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Whether debug logs should be displayed",
		},
	},
	Action: createAction,
}

func createAction(c *cli.Context) (err error) {
	var (
		name  = c.String("name")
		size  = c.String("size")
		root  = c.String("root")
		inode = c.Uint64("inode")
		debug = c.Bool("debug")

		sizeInBytes uint64
	)

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if name == "" || size == "" || root == "" {
		cli.ShowCommandHelp(c, "create")
		err = cli.NewExitError(
			"Name, size and root are required parameters.", 1)
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

	sizeInBytes, err = manager.FromHumanSize(size)
	if err != nil {
		err = cli.NewExitError(errors.Wrapf(err,
			"Size '%s' can't be converted to uint64 bytes", size), 1)
		return
	}

	_, err = mgr.Create(manager.Volume{
		Name:  name,
		Size:  sizeInBytes,
		INode: inode,
	})
	if err != nil {
		err = cli.NewExitError(errors.Wrapf(err,
			"Couldn't create volume name=%s bytes=%d inode=%d",
			name, sizeInBytes, inode), 1)
		return
	}

	return
}
