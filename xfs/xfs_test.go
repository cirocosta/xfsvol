package xfs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cirocosta/xfsvol/xfs"
	"github.com/stretchr/testify/assert"
)

// setupTestFs takes a filesystem description as
// a variable and setups the desired structure under
// a temp directory.
func setupTestFs(fs []string) (root string, err error) {
	var (
		extension string
		directory string
	)

	root, err = ioutil.TempDir("", "")
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

func makeBigString(size int) (res string) {
	var buffer = make([]byte, size)
	for ndx := range buffer {
		buffer[ndx] = 'a'
	}

	res = string(buffer)
	return
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
			root, err = setupTestFs(tc.fs)
			assert.NoError(t, err)
			defer os.RemoveAll(root)

			err = xfs.MakeBackingFsDev(filepath.Join(root, tc.root), tc.file)
			if tc.shouldFail {
				assert.Error(t, err)
			}
		})
	}
}
