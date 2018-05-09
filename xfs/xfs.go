package xfs

// #include "./xfs.h"
import "C"

import (
	"unsafe"

	"github.com/pkg/errors"
)

// Quota defines the limit params to be applied or that
// are already set to a project:
// -	Size:	number of blk sizes that can be
//		commited
// -	INode:	maximum number of INodes that
//		can be created
type Quota struct {
	Size  uint64
	INode uint64
}

// SetProjectQuota sets quota settings associated with a given
// projectId controlled by a given block device.
func SetProjectQuota(blockDevice string, projectId uint32, q *Quota) (err error) {
	if blockDevice == "" {
		err = errors.Errorf("blockDevice must be specified")
		return
	}

	var (
		blockDeviceString = C.CString(blockDevice)
		quota             = &C.struct_xfs_quota{
			inodes: C.__u64(q.INode),
			size:   C.__u64(q.Size),
		}
	)
	defer C.free(unsafe.Pointer(blockDeviceString))

	ret, err := C.xfs_set_project_quota(blockDeviceString,
		C.__u32(projectId),
		quota)
	if ret == -1 {
		err = errors.Wrapf(err,
			"failed to set project quota "+
				"prj=%d dev=%d quota-size=%d quota-inodes=%d",
			projectId, blockDevice, q.Size, q.INode)
		return
	}

	return
}

// GetProjectQuota retrieves the quota settings associated
// with a project-id controlled by a given block device.
func GetProjectQuota(blockDevice string, projectId uint32) (q *Quota, err error) {
	if blockDevice == "" {
		err = errors.Errorf("blockDevice must be specified")
		return
	}

	var (
		blockDeviceString = C.CString(blockDevice)
		quota             = new(C.struct_xfs_quota)
	)
	defer C.free(unsafe.Pointer(blockDeviceString))

	ret, err := C.xfs_get_project_quota(blockDeviceString,
		C.__u32(projectId),
		quota)
	if ret == -1 {
		err = errors.Wrapf(err,
			"failed to retrieve project quota - prj=%d dev=%s",
			projectId, blockDevice)
		return
	}

	q = new(Quota)
	q.INode = uint64(quota.inodes)
	q.Size = uint64(quota.size)

	return
}

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

// SetProjectId sets the value of the extended attribute projectid associated
// with a given directory, as well as setting necessary flags (PROJINHERIT).
func SetProjectId(directory string, projectId uint32) (err error) {
	if directory == "" {
		err = errors.Errorf("directory must be specified")
		return
	}

	var directoryString = C.CString(directory)
	defer C.free(unsafe.Pointer(directoryString))

	ret, err := C.xfs_set_project_id(directoryString, C.__u32(projectId))
	if ret == -1 {
		err = errors.Wrapf(err,
			"failed to set project-id %d to directory %s",
			projectId, directory)
		return
	}

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
