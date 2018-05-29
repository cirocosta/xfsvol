package quota_test

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/cirocosta/xfsvol/quota"
	"github.com/cirocosta/xfsvol/quota/xfs"
	"github.com/stretchr/testify/assert"

	utils "github.com/cirocosta/xfsvol/test_utils"
)

var (
	controlTypesMounts = map[string]string{
		"xfs":  "/mnt/xfs/tmp",
		"ext4": "/mnt/ext4/tmp",
	}
	errNotImplemented = errors.Errorf("control type not implemented")
)

// createControl is a factory that takes a controlType (either
// `xfs` or `ext4`) and builds a Control implementor based on
// the type of configuration passed.
func createControl(controlType, root string) (control Control, err error) {
	switch controlType {
	case "xfs":
		control, err = xfs.NewControl(xfs.ControlConfig{
			Dir: root,
		})
		return
	case "ext4":
		err = errNotImplemented
		return
	default:
		err = errors.Errorf("unknown control type %s",
			controlType)
		return
	}

	return
}

func TestControlImplementations(t *testing.T) {
	var (
		dir     string
		err     error
		control quota.Control
	)

	for controlType, controlMount := range controlTypesMounts {
		t.Run(controlType+" succeeds if properly mounted directory", func(t *testing.T) {
			dir, err := ioutil.TempDir(xfsMount, "")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			_, err = quota.NewControl(quota.ControlConfig{
				BasePath: dir,
			})
			assert.NoError(t, err)
		})

		t.Run(controlType+" quota assignment fails if outside tree", func(t *testing.T) {
			dirOutside, err := ioutil.TempDir("", "")
			assert.NoError(t, err)
			defer os.RemoveAll(dirOutside)

			dir, err := ioutil.TempDir(xfsMount, "")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			ctl, err := quota.NewControl(quota.ControlConfig{
				BasePath: dir,
			})
			assert.NoError(t, err)

			err = ctl.SetQuota(dirOutside, quota.Quota{
				Size: 10 * (1 << 20),
			})
			assert.Error(t, err)
		})

		t.Run(controlType+" fails to assign to inexistent dir in tree", func(t *testing.T) {

			dir, err := ioutil.TempDir(xfsMount, "")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			dirInside := path.Join(dir, "abc")

			ctl, err := quota.NewControl(quota.ControlConfig{
				BasePath: dir,
			})
			assert.NoError(t, err)

			err = ctl.SetQuota(dirInside, quota.Quota{
				Size: 10 * (1 << 20),
			})
			assert.Error(t, err)
		})

		t.Run(controlType+" succeeds to assign to dir in tree", func(t *testing.T) {

			dir, err := ioutil.TempDir(xfsMount, "")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			dirInside := path.Join(dir, "abc")
			err = os.MkdirAll(dirInside, 0755)
			assert.NoError(t, err)

			ctl, err := quota.NewControl(quota.ControlConfig{
				BasePath: dir,
			})
			assert.NoError(t, err)

			err = ctl.SetQuota(dirInside, quota.Quota{
				Size: 10 * (1 << 20),
			})
			assert.NoError(t, err)

		})

		t.Run(controlType+" flatly enforces disk quota", func(t *testing.T) {

			dir, err := ioutil.TempDir(xfsMount, "")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			dir1M := path.Join(dir, "1M")
			dir2M := path.Join(dir, "2M")

			assert.NoError(t, os.MkdirAll(dir1M, 0755))
			assert.NoError(t, os.MkdirAll(dir2M, 0755))

			var startingProjectId uint32 = 100
			ctl, err := quota.NewControl(quota.ControlConfig{
				BasePath:          dir,
				StartingProjectId: &startingProjectId,
			})
			assert.NoError(t, err)

			assert.NoError(t, ctl.SetQuota(dir1M, quota.Quota{
				Size: 1 * (1 << 20),
			}))

			assert.NoError(t, ctl.SetQuota(dir2M, quota.Quota{
				Size: 1 * (2 << 20),
			}))

			fileDir1M, err := os.Create(path.Join(dir1M, "file"))
			assert.NoError(t, err)

			fileDir2M, err := os.Create(path.Join(dir2M, "file"))
			assert.NoError(t, err)

			assert.Error(t, utils.WriteBytes(fileDir1M, 'c', 2*(1<<20)))
			assert.NoError(t, utils.WriteBytes(fileDir2M, 'c', 1*(1<<20)))

		})

		t.Run(controlType+" flatly enforces inode quota", func(t *testing.T) {
			dir, err := ioutil.TempDir(xfsMount, "aa")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			dirA := path.Join(dir, "A")
			dirB := path.Join(dir, "B")

			assert.NoError(t, os.MkdirAll(dirA, 0755))
			assert.NoError(t, os.MkdirAll(dirB, 0755))

			ctl, err := quota.NewControl(quota.ControlConfig{
				BasePath: dir,
			})
			assert.NoError(t, err)

			assert.NoError(t, ctl.SetQuota(dirA, quota.Quota{
				Size:  2 * (1 << 20),
				INode: 30,
			}))

			assert.NoError(t, ctl.SetQuota(dirB, quota.Quota{
				Size:  2 * (1 << 20),
				INode: 300,
			}))

			assert.Error(t, utils.CreateFiles(dirA, 100))
			assert.NoError(t, utils.CreateFiles(dirB, 100))

		})
	}
}
