package xfs

// #include "./xfs.h"
import "C"

import (
	"unsafe"

	"github.com/pkg/errors"
)

// GetProjectId retrieves the extended attribute projectid associated
// with a given directory.
func GetProjectId(directory string) (projectId uint32, err error) {
	if directory == "" {
		err = errors.Errorf("directory must be specified")
		return
	}

	var directoryString = C.CString(directory)
	defer C.free(unsafe.Pointer(directoryString))

	ret, err := C.xfs_get_project_id(directoryString)
	if ret == -1 {
		err = errors.Wrapf(err,
			"failed to get project-id from directory %s",
			directory)
		return
	}

	projectId = uint32(ret)
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
