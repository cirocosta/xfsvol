package quota_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cirocosta/xfsvol/quota"
	"github.com/stretchr/testify/assert"

	utils "github.com/cirocosta/xfsvol/test_utils"
)

const (
	// xfsMountPath corresponds to the project-quota-enabled
	// mount path that must be set up before running the
	// tests.
	//
	// Check `.travis/setup.sh` for some more information on
	// how to have this properly done.
	xfsMountPath = "/mnt/xfs"

	// xfsMountPathWithoutQuota specifies the test mount path
	// of a device that has xfs properly set up but without
	// project quota support.
	//
	// Check `.travis/setup.sh` for some more information on
	// how to have this properly done.
	xfsMountPathWithoutQuota = "/mnt/xfs-without-quota"
)

// setupTestFs takes a filesystem description as
// a variable and setups the desired structure under
// a temp directory.
func setupTestFs(base string, fs []string) (root string, err error) {
	var (
		extension string
		directory string
	)

	root, err = ioutil.TempDir(base, "")
	if err != nil {
		return
	}

	for _, element := range fs {
		element = filepath.Join(root, element)
		extension = filepath.Ext(element)
		directory = filepath.Dir(element)

		err = os.MkdirAll(directory, 0755)
		if err != nil && !os.IsExist(err) {
			return
		}

		if extension != "" {
			_, err = os.Create(element)
		} else {
			err = os.MkdirAll(element, 0755)
		}
		if err != nil {
			return
		}
	}

	return
}

// makeBigString creates a string filled with a single
// character that is `size` big.
func makeBigString(size int) (res string) {
	var buffer = make([]byte, size)
	for ndx := range buffer {
		buffer[ndx] = 'a'
	}

	res = string(buffer)
	return
}

func TestGetProjectId(t *testing.T) {
	var testCases = []struct {
		desc        string
		projectId   uint32
		fs          []string
		target      string
		shouldError bool
	}{
		{
			desc:        "fails if not a directory",
			fs:          []string{"/file.txt"},
			target:      "file.txt",
			shouldError: true,
		},
		{
			desc:        "fails if directory doesnt exist",
			fs:          []string{"/"},
			target:      "dir",
			shouldError: true,
		},
		{
			desc:        "returns zero if no attributes associated",
			fs:          []string{"/dir"},
			target:      "dir",
			shouldError: false,
			projectId:   0,
		},
	}

	var (
		err       error
		root      string
		projectId uint32
	)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			root, err = setupTestFs(xfsMountPath, tc.fs)
			assert.NoError(t, err)
			defer os.RemoveAll(root)

			projectId, err = quota.GetProjectId(filepath.Join(root, tc.target))
			if tc.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.projectId, projectId)
		})
	}

}

func TestSetProjectId(t *testing.T) {
	var testCases = []struct {
		desc        string
		projectId   uint32
		fs          []string
		fsBasePath  string
		target      string
		shouldError bool
		basePath    string
	}{
		{
			desc:        "fails if not a directory",
			fs:          []string{"/file.txt"},
			fsBasePath:  filepath.Join(xfsMountPath, "/tmp"),
			target:      "file.txt",
			shouldError: true,
		},
		{
			desc:        "fails if directory doesnt exist",
			fs:          []string{"/"},
			fsBasePath:  filepath.Join(xfsMountPath, "/tmp"),
			target:      "dir",
			shouldError: true,
		},
		{
			desc:        "fails if not an xfs filesystem",
			fs:          []string{"/dir"},
			target:      "dir",
			fsBasePath:  "/tmp",
			projectId:   123,
			shouldError: true,
		},
		{
			desc:        "associates projectid with directory in xfs fs",
			fs:          []string{"/dir"},
			target:      "dir",
			fsBasePath:  filepath.Join(xfsMountPath, "/tmp"),
			projectId:   123,
			shouldError: false,
		},
	}

	var (
		err       error
		root      string
		projectId uint32
	)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			root, err = setupTestFs(tc.fsBasePath, tc.fs)
			assert.NoError(t, err)
			defer os.RemoveAll(root)

			err = quota.SetProjectId(filepath.Join(root, tc.target), tc.projectId)
			if tc.shouldError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			projectId, err = quota.GetProjectId(filepath.Join(root, tc.target))
			assert.NoError(t, err)
			assert.Equal(t, tc.projectId, projectId)
		})
	}
}

func TestSetProjectId_childrenHaveProjectIdSet(t *testing.T) {
	const desiredProjectId uint32 = 543

	var (
		root      string
		err       error
		projectId uint32
	)

	root, err = setupTestFs(
		filepath.Join(xfsMountPath, "/tmp"),
		[]string{"/dir"})
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	err = quota.SetProjectId(filepath.Join(root, "/dir"), desiredProjectId)
	assert.NoError(t, err)

	err = os.MkdirAll(filepath.Join(root, "/dir", "/child"), 0755)
	assert.NoError(t, err)

	projectId, err = quota.GetProjectId(filepath.Join(root, "/dir"))
	assert.NoError(t, err)
	assert.Equal(t, desiredProjectId, projectId)

	projectId, err = quota.GetProjectId(filepath.Join(root, "/dir", "/child"))
	assert.NoError(t, err)
	assert.Equal(t, desiredProjectId, projectId)
}

func TestMakeBackingFsDev(t *testing.T) {
	var testCases = []struct {
		desc       string
		root       string
		file       string
		fs         []string
		shouldFail bool
	}{
		{
			desc:       "fails if root doesnt exist",
			root:       "/dir",
			file:       "dev",
			fs:         []string{},
			shouldFail: true,
		},
		{
			desc: "fails if root is not a dir",
			root: "/root.txt",
			file: "dev",
			fs: []string{
				"/root.txt",
			},
			shouldFail: true,
		},
		{
			desc: "fails if file name is too long",
			root: "/",
			file: makeBigString(1 << 20),
			fs: []string{
				"/",
			},
			shouldFail: true,
		},
		{
			desc: "succeeds if root exists and is a dir",
			root: "/dir",
			file: "dev",
			fs: []string{
				"/dir",
			},
			shouldFail: false,
		},
	}

	var (
		root string
		err  error
	)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			root, err = setupTestFs("", tc.fs)
			assert.NoError(t, err)
			defer os.RemoveAll(root)

			err = quota.MakeBackingFsDev(filepath.Join(root, tc.root), tc.file)
			if tc.shouldFail {
				assert.Error(t, err)
			}
		})
	}
}

func TestMakeBackingFsDev_succeedsIfAlreadyExists(t *testing.T) {
	root, err := setupTestFs("", []string{"/dir"})
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	err = quota.MakeBackingFsDev(filepath.Join(root, "/dir"), "device")
	assert.NoError(t, err)

	err = quota.MakeBackingFsDev(filepath.Join(root, "/dir"), "device")
	assert.NoError(t, err)

	err = quota.MakeBackingFsDev(filepath.Join(root, "/dir"), "device")
	assert.NoError(t, err)
}

func TestSetProjectQuota_failsIfBlockDeviceDoesntExist(t *testing.T) {
	var (
		fs                           = []string{"/dir"}
		inexistentProjectId   uint32 = 999
		inexistentBlockDevice string
	)

	root, err := setupTestFs(xfsMountPath, fs)
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	inexistentBlockDevice = filepath.Join(root, "inexistent-device")

	err = quota.SetProjectQuota(inexistentBlockDevice, inexistentProjectId, &quota.Quota{
		Size:  123,
		INode: 123,
	})
	assert.Error(t, err)
}

func TestSetProjectQuota_succeeds(t *testing.T) {
	var (
		fs                   = []string{"/dir"}
		projectId     uint32 = 999
		expectedQuota        = &quota.Quota{Size: 1 << 20, INode: 1 << 20}
		actualQuota   *quota.Quota
		blockDevice   string
	)

	root, err := setupTestFs(xfsMountPath, fs)
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	blockDevice = filepath.Join(root, "block-device")
	err = quota.MakeBackingFsDev(root, "block-device")
	assert.NoError(t, err)

	err = quota.SetProjectId(filepath.Join(root, "dir"), projectId)
	assert.NoError(t, err)

	err = quota.SetProjectQuota(blockDevice, projectId, expectedQuota)
	assert.NoError(t, err)

	actualQuota, err = quota.GetProjectQuota(blockDevice, projectId)
	assert.NoError(t, err)

	assert.Equal(t, expectedQuota.Size, actualQuota.Size)
	assert.Equal(t, expectedQuota.INode, actualQuota.INode)
}

func TestGetProjectStats(t *testing.T) {
	root, err := setupTestFs(xfsMountPath, []string{"/dir"})
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	var (
		blockDevice        = filepath.Join(root, "block-device")
		directory          = filepath.Join(root, "/dir")
		file               = filepath.Join(directory, "file")
		projectId   uint32 = 333
	)

	err = quota.MakeBackingFsDev(root, "block-device")
	assert.NoError(t, err)

	err = quota.SetProjectId(directory, projectId)
	assert.NoError(t, err)

	err = quota.SetProjectQuota(blockDevice, projectId, &quota.Quota{})
	assert.NoError(t, err)

	quota1, err := quota.GetProjectQuota(blockDevice, projectId)
	assert.NoError(t, err)

	err = utils.CreateFiles(directory, 100)
	assert.NoError(t, err)

	file1M, err := os.Create(file)
	assert.NoError(t, err)

	err = utils.WriteBytes(file1M, 'c', 1<<20)
	assert.NoError(t, err)

	file1M.Sync()

	quota2, err := quota.GetProjectQuota(blockDevice, projectId)
	assert.NoError(t, err)

	assert.Equal(t, uint64(101), quota2.UsedInode-quota1.UsedInode)
	assert.True(t, quota1.UsedSize < (1<<15))
	assert.True(t, quota2.UsedSize > (1<<20))
}

func TestIsQuotaEnabled_failsIfBlockDeviceDoesntExist(t *testing.T) {
	_, err := quota.IsQuotaEnabled("/inexistent-block/_device")
	assert.Error(t, err)
}

func TestIsQuotaEnabled_notEnabledIfNotXfs(t *testing.T) {
	root, err := setupTestFs("", []string{"/"})
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	err = quota.MakeBackingFsDev(root, "block-device")
	assert.NoError(t, err)

	_, err = quota.IsQuotaEnabled(filepath.Join(root, "block-device"))
	assert.NoError(t, err)
}

func TestIsQuotaEnabled_notEnabledIfNoProjectQuotaSet(t *testing.T) {
	root, err := setupTestFs(xfsMountPathWithoutQuota, []string{"/"})
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	err = quota.MakeBackingFsDev(root, "block-device")
	assert.NoError(t, err)

	isEnabled, err := quota.IsQuotaEnabled(filepath.Join(root, "block-device"))
	assert.NoError(t, err)
	assert.False(t, isEnabled)
}

func TestIsQuotaEnabled(t *testing.T) {
	root, err := setupTestFs(xfsMountPath, []string{"/"})
	assert.NoError(t, err)
	defer os.RemoveAll(root)

	err = quota.MakeBackingFsDev(root, "block-device")
	assert.NoError(t, err)

	isEnabled, err := quota.IsQuotaEnabled(filepath.Join(root, "block-device"))
	assert.NoError(t, err)
	assert.True(t, isEnabled)
}
