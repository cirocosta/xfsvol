package xfs

// #include "./xfs.h"
import "C"

import (
	"unsafe"

	"github.com/pkg/errors"
)

func GetProjectId(directory string) (projectId uint32, err error) {
	return
}

// MakeBackingFsDev creates a block device under the directory
// specified in the `root` argument.
func MakeBackingFsDev(root, file string) (err error) {
	if root == "" || file == "" {
		err = errors.Errorf("root and file must be provided")
		return
	}

	var (
		rootString = C.CString(root)
		fileString = C.CString(file)
	)

	defer C.free(unsafe.Pointer(rootString))
	defer C.free(unsafe.Pointer(fileString))

	ret, err := C.xfs_create_fs_block_dev(rootString, fileString)
	if ret == -1 {
		err = errors.Wrapf(err,
			"failed to create fs block device %s/%s",
			root, file)
		return
	}

	return
}
