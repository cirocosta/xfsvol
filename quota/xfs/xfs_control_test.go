package xfs_test

import (
	"testing"

	"github.com/cirocosta/xfsvol/quota/xfs"
	"github.com/stretchr/testify/assert"
)

// TODO implementation-defined
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
