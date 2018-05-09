package manager_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/cirocosta/xfsvol/manager"
	"github.com/stretchr/testify/assert"

	utils "github.com/cirocosta/xfsvol/test_utils"
)

const (
	xfsMount = "/mnt/xfs/tmp"
)

func TestNew_failsWithoutRootSpecified(t *testing.T) {
	_, err := manager.New(manager.Config{})
	assert.Error(t, err)
}

func TestNew_failsWithInexistentRoot(t *testing.T) {
	_, err := manager.New(manager.Config{
		Root: "/a/b/c/d/e/f/g/h/i",
	})
	assert.Error(t, err)
}

func TestNew_failsWithNonAbsolutePath(t *testing.T) {
	_, err := manager.New(manager.Config{
		Root: "var/log",
	})
	assert.Error(t, err)
}

func TestNew_succeedsWithWriteableAbsolutePath(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)
}

func TestCreate_failsIfEmptyPath(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	_, err = m.Create(manager.Volume{})
	assert.Error(t, err)
}

func TestCreate_failsWithWeirdCharacters(t *testing.T) {
	var weirdPaths = []string{
		"./",
		"'aa",
		"bb+",
		"a b c",
	}

	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	for _, path := range weirdPaths {
		_, err := m.Create(manager.Volume{Name: path})
		assert.Error(t, err)
	}
}

func TestCreate_succeedsWithNormalPathAndSize(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	absPath, err := m.Create(manager.Volume{
		Name: "abc",
		Size: manager.MustFromHumanSize("10M"),
	})
	assert.NoError(t, err)
	assert.Equal(t, path.Join(dir, "abc"), absPath)

	finfo, err := os.Stat(absPath)
	assert.NoError(t, err)
	assert.True(t, finfo.IsDir())
}

func TestCreate_actuallyHasQuotaEnforced(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	absPath, err := m.Create(manager.Volume{
		Name: "abc",
		Size: manager.MustFromHumanSize("10M"),
	})
	assert.NoError(t, err)

	filePath := path.Join(absPath, "12Mfile")
	fd, err := os.Create(filePath)
	assert.NoError(t, err)
	defer fd.Close()

	fmt.Println(filePath)
	assert.Error(t, utils.WriteBytes(fd, 'c', 20*(1<<20)))
}

func TestList_canList0Directorise(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	dirs, err := m.List()
	assert.NoError(t, err)
	assert.Len(t, dirs, 0)
}

func TestCreate_cantCreateWithEmptySize(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	_, err = m.Create(manager.Volume{
		Name: "abc",
		Size: 0,
	})
	assert.Error(t, err)
}

func TestList_listsDirectories(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	_, err = m.Create(manager.Volume{
		Name: "abc",
		Size: manager.MustFromHumanSize("10M"),
	})
	assert.NoError(t, err)

	_, err = m.Create(manager.Volume{
		Name: "def",
		Size: manager.MustFromHumanSize("10M"),
	})
	assert.NoError(t, err)

	dirs, err := m.List()
	assert.NoError(t, err)
	assert.Len(t, dirs, 2)
	assert.Equal(t, "abc", dirs[0].Name)
	assert.Equal(t, "10MB", manager.HumanSize(dirs[0].Size))
	assert.Equal(t, "def", dirs[1].Name)
	assert.Equal(t, "10MB", manager.HumanSize(dirs[1].Size))
}

func TestGet_doesntErrorIfNotFound(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	_, found, err := m.Get("abc")
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestGet_findsDirectory(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	_, err = m.Create(manager.Volume{
		Name: "abc",
		Size: manager.MustFromHumanSize("10 MB"),
	})
	assert.NoError(t, err)

	vol, found, err := m.Get("abc")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, path.Join(dir, "abc"), vol.Path)
}

func TestDelete_succeedsForExistentVolume(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	absPath, err := m.Create(manager.Volume{
		Name: "abc",
		Size: manager.MustFromHumanSize("10 MB"),
	})
	assert.NoError(t, err)
	assert.Equal(t, path.Join(dir, "abc"), absPath)

	finfo, err := os.Stat(absPath)
	assert.NoError(t, err)
	assert.True(t, finfo.IsDir())

	err = m.Delete("abc")
	assert.NoError(t, err)

	finfo, err = os.Stat(absPath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestDelete_failsForInexistentVolume(t *testing.T) {
	dir, err := ioutil.TempDir(xfsMount, "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	m, err := manager.New(manager.Config{
		Root: dir,
	})
	assert.NoError(t, err)

	err = m.Delete("abc")
	assert.Error(t, err)
}
