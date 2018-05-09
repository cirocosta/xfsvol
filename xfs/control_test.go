package xfs_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/cirocosta/xfsvol/xfs"
	"github.com/stretchr/testify/assert"

	utils "github.com/cirocosta/xfsvol/test_utils"
)

const xfsMount = "/mnt/xfs/tmp"

func TestControl_failsIfNotXfsDirectory(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = xfs.NewControl(xfs.ControlConfig{
		BasePath: dir,
	})
	assert.Error(t, err)
}

func TestControl_succeedsIfXFSBasedDirectory(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = xfs.NewControl(xfs.ControlConfig{
		BasePath: dir,
	})
	assert.NoError(t, err)
}

func TestControl_createsBackingFsBlockDev(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ctl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath: dir,
	})
	assert.NoError(t, err)

	finfo, err := os.Stat(ctl.GetBackingFsBlockDev())
	assert.NoError(t, err)
	assert.False(t, finfo.Mode().IsRegular())
	assert.False(t, finfo.Mode().IsDir())
	assert.True(t, finfo.Mode()&os.ModeDevice != 0)
}

func TestControl_quotaAssignmentFailsToIfDirectoryOutsideTree(t *testing.T) {
	dirOutside, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dirOutside)

	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ctl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath: dir,
	})
	assert.NoError(t, err)

	err = ctl.SetQuota(dirOutside, xfs.Quota{
		Size: 10 * (1 << 20),
	})
	assert.Error(t, err)
}

func TestControl_failsToAssignQuotaToInexistentDirectoryWithinTree(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	dirInside := path.Join(dir, "abc")

	ctl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath: dir,
	})
	assert.NoError(t, err)

	err = ctl.SetQuota(dirInside, xfs.Quota{
		Size: 10 * (1 << 20),
	})
	assert.Error(t, err)
}

func TestControl_succeedsToAssignQuotaToDirectoryWithinTree(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	dirInside := path.Join(dir, "abc")
	err = os.MkdirAll(dirInside, 0755)
	assert.NoError(t, err)

	ctl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath: dir,
	})
	assert.NoError(t, err)

	err = ctl.SetQuota(dirInside, xfs.Quota{
		Size: 10 * (1 << 20),
	})
	assert.NoError(t, err)
}

func TestControl_flatlyEnforcesDiskQuota(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	dir1M := path.Join(dir, "1M")
	dir2M := path.Join(dir, "2M")

	assert.NoError(t, os.MkdirAll(dir1M, 0755))
	assert.NoError(t, os.MkdirAll(dir2M, 0755))

	var startingProjectId uint32 = 100
	ctl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath:          dir,
		StartingProjectId: &startingProjectId,
	})
	assert.NoError(t, err)

	assert.NoError(t, ctl.SetQuota(dir1M, xfs.Quota{
		Size: 1 * (1 << 20),
	}))

	assert.NoError(t, ctl.SetQuota(dir2M, xfs.Quota{
		Size: 1 * (2 << 20),
	}))

	fileDir1M, err := os.Create(path.Join(dir1M, "file"))
	assert.NoError(t, err)

	fileDir2M, err := os.Create(path.Join(dir2M, "file"))
	assert.NoError(t, err)

	assert.Error(t, utils.WriteBytes(fileDir1M, 'c', 2*(1<<20)))
	assert.NoError(t, utils.WriteBytes(fileDir2M, 'c', 1*(1<<20)))
}

func TestControl_flatlyEnforcesINodeQuota(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "aa")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	dirA := path.Join(dir, "A")
	dirB := path.Join(dir, "B")

	assert.NoError(t, os.MkdirAll(dirA, 0755))
	assert.NoError(t, os.MkdirAll(dirB, 0755))

	ctl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath: dir,
	})
	assert.NoError(t, err)

	assert.NoError(t, ctl.SetQuota(dirA, xfs.Quota{
		Size:  2 * (1 << 20),
		INode: 30,
	}))

	assert.NoError(t, ctl.SetQuota(dirB, xfs.Quota{
		Size:  2 * (1 << 20),
		INode: 300,
	}))

	assert.Error(t, utils.CreateFiles(dirA, 100))
	assert.NoError(t, utils.CreateFiles(dirB, 100))
}
