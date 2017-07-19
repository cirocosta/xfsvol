package lib

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const XfsMount = "/mnt/xfs/tmp"

func TestControl_failsIfNotXfsDirectory(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = NewControl(dir)
	assert.Error(t, err)
}

func TestControl_succeedsIfXFSBasedDirectory(t *testing.T) {
	dir, err := ioutil.TempDir(XfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = NewControl(dir)
	assert.NoError(t, err)
}

func TestControl_createsBackingFsBlockDev(t *testing.T) {
	dir, err := ioutil.TempDir(XfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ctl, err := NewControl(dir)
	assert.NoError(t, err)

	finfo, err := os.Stat(ctl.backingFsBlockDev)
	assert.NoError(t, err)
	assert.False(t, finfo.Mode().IsRegular())
	assert.False(t, finfo.Mode().IsDir())
	assert.True(t, finfo.Mode()&os.ModeDevice != 0)
}

func TestControl_quotaAssignmentFailsToIfDirectoryOutsideTree(t *testing.T) {
	dirOutside, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dirOutside)

	dir, err := ioutil.TempDir(XfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ctl, err := NewControl(dir)
	assert.NoError(t, err)

	err = ctl.SetQuota(dirOutside, Quota{
		Size: 10 * (1 << 20),
	})
	assert.Error(t, err)
}

func TestControl_failsToAssignQuotaToInexistentDirectoryWithinTree(t *testing.T) {
	dir, err := ioutil.TempDir(XfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	dirInside := path.Join(dir, "abc")

	ctl, err := NewControl(dir)
	assert.NoError(t, err)

	err = ctl.SetQuota(dirInside, Quota{
		Size: 10 * (1 << 20),
	})
	assert.Error(t, err)
}

func TestControl_succeedsToAssignQuotaToDirectoryWithinTree(t *testing.T) {
	dir, err := ioutil.TempDir(XfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	dirInside := path.Join(dir, "abc")
	err = os.MkdirAll(dirInside, 0755)
	assert.NoError(t, err)

	ctl, err := NewControl(dir)
	assert.NoError(t, err)

	err = ctl.SetQuota(dirInside, Quota{
		Size: 10 * (1 << 20),
	})
	assert.NoError(t, err)
}
