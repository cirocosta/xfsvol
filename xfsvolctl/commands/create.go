package commands

import (
	"github.com/cirocosta/xfsvol/manager"
	"github.com/pkg/errors"
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
	},
	Action: createAction,
}

func createAction(c *cli.Context) (err error) {
	var (
		name  = c.String("name")
		size  = c.String("size")
		root  = c.String("root")
		inode = c.Uint64("inode")

		sizeInBytes uint64
	)

	if name == "" || size == "" || root == "" {
		cli.ShowCommandHelp(c, "create")
		err = errors.Errorf("Name, Size and Root are required parameters.")
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

	sizeInBytes, err = manager.FromHumanSize(size)
	if err != nil {
		err = errors.Wrapf(err,
			"Size supplied by user '%s' can't be converted to uint64 bytes",
			size)
		return
	}

	_, err = mgr.Create(manager.Volume{
		Name:  name,
		Size:  sizeInBytes,
		INode: inode,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't create volume name=%s bytes=%d inode=%d",
			name, sizeInBytes, inode)
		return
	}

	return
}
