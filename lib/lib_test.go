package lib

import (
	"io/ioutil"
	"os"
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
