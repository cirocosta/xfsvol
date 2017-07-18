package commands

import (
  "github.com/pkg/errors"
  "github.com/cirocosta/xfsvol/manager"
	"gopkg.in/urfave/cli.v1"

  log "github.com/sirupsen/logrus"
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

      1. create a volume with an upper limit of 10M:

         lsblk
         NAME   MAJ:MIN RM   SIZE RO TYPE MOUNTPOINT
         loop0    7:0    0     2G  0 loop /mnt/xfs
         
         mount | grep xfs
         /dev/loop0 on /mnt/xfs type xfs (rw,relatime,attr2,inode64,prjquota)
         
         xfsvolctl create --root /mnt/xfs --name myvol --size 10M

    Note:
      -  In order to have the creation functioning you must
         first have a mount point in the filesystem that
         is mounted on top of XFS and has 'pquota' set.
    `,
  Flags: []cli.Flag{
    cli.StringFlag{
      Name: "name, n",
      Usage: "Name of the volume to create",
    },
    cli.StringFlag{
      Name: "size, s",
      Usage: "Size of the XFS project quota to apply",
    },
    cli.StringFlag{
      Name: "root, r",
      Usage: "Root of the volume creation",
    },
  },
	Action: func (c *cli.Context) (err error) {
    var name = c.String("name")
    var size = c.String("size")
    var root = c.String("root")

    if name == "" || size == "" || root == "" {
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

    sizeInBytes, err := manager.FromHumanSize(size)
    if err != nil {
      err = errors.Wrapf(err,
        "Size supplied by user can't be converted to uint64 bytes")
      return
    }

    location, err := mgr.Create(manager.Volume{
      Name: name,
      Size: sizeInBytes,
    })
    if err != nil {
      err = errors.Wrapf(err,
        "Couldn't create volume name=%s bytes=%d",
        name, sizeInBytes)
      return
    }

    log.WithField("location", location).Info("volume created")
		return
	},
}
