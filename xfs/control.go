// +build linux
// xfs implements XFS project quota controls for setting quota limits
// on a newly created directory.
package xfs

// #include "./control.h"
import "C"

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"
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

// Control gives the context to be used by storage driver
// who wants to apply project quotas to container dirs.
type Control struct {
	backingFsBlockDev string
	nextProjectID     uint32
	quotas            map[string]uint32
	logger            zerolog.Logger
}

// ControlConfig specifies the configuration to be used by
// the controller that will hold the quota allocation state.
type ControlConfig struct {
	StartingProjectId *uint32
	BasePath          string
}

// NewControl initializes project quota support.
//
// It first tests to make sure that quota can be set on a test dir
// and find the first project id to be used for the next container creation.
//
// Returns nil (and error) if project quota is not supported.
//
// First get the project id of the home directory.
// This test will fail if the backing fs is not xfs.
//
// xfs_quota tool can be used to assign a project id to the driver home directory, e.g.:
//
//    echo 999:/var/lib/docker/overlay2 >> /etc/projects
//    echo docker:999 >> /etc/projid
//    xfs_quota -x -c 'project -s docker' /<xfs mount point>
//
// In that case, the home directory project id will be used as a "start offset"
// and all containers will be assigned larger project ids (e.g. >= 1000).
// This is a way to prevent xfs_quota management from conflicting with docker.
//
// Then try to create a test directory with the next project id and set a quota
// on it. If that works, continue to scan existing containers to map allocated
// project ids.
//
func NewControl(cfg ControlConfig) (c Control, err error) {
	var minProjectID uint32

	if cfg.BasePath == "" {
		err = errors.Errorf("BasePath must be provided")
		return
	}

	c.logger = zerolog.New(os.Stdout).With().Str("from", "control").Logger()

	if cfg.StartingProjectId == nil {
		minProjectID, err = getProjectID(cfg.BasePath)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to retrieve projectid from basepath %s",
				cfg.BasePath)
			return
		}

		minProjectID++
	} else {
		minProjectID = *cfg.StartingProjectId
	}

	//
	// create backing filesystem device node
	//
	c.backingFsBlockDev, err = makeBackingFsDev(cfg.BasePath)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create backingfs dev for base path %s",
			cfg.BasePath)
		return
	}

	//
	// Test if filesystem supports project quotas by trying to set
	// a quota on the first available project id
	//
	err = setProjectQuota(c.backingFsBlockDev, minProjectID, Quota{
		Size: 0,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"failed to set quota on the first available"+
				" proj id after base path %s",
			cfg.BasePath)
		return
	}

	c.nextProjectID = minProjectID + 1
	c.quotas = make(map[string]uint32)

	//
	// get first project id to be used for next container
	//
	err = c.findNextProjectID(cfg.BasePath)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to find next projectId from basepath %s", cfg.BasePath)
		return
	}

	c.logger.Debug().
		Str("base-path", cfg.BasePath).
		Uint32("next-project-id", c.nextProjectID).
		Msg("new control created")

	return
}

// GetBackingFsBlockDev retrieves the backing block device
// configured for the current quota control instance.
func (c *Control) GetBackingFsBlockDev() (blockDev string) {
	blockDev = c.backingFsBlockDev
	return
}

// SetQuota assigns a unique project id to directory and set the
// quota for that project id.
func (q *Control) SetQuota(targetPath string, quota Quota) (err error) {
	projectID, ok := q.quotas[targetPath]
	if !ok {
		projectID = q.nextProjectID

		//
		// assign project id to new container directory
		//
		err = setProjectID(targetPath, projectID)
		if err != nil {
			err = errors.Wrapf(err, "couldn't set project id")
			return
		}

		q.quotas[targetPath] = projectID
		q.nextProjectID++
	}

	q.logger.Debug().
		Uint32("project-id", projectID).
		Str("target-path", targetPath).
		Uint64("quota-size", quota.Size).
		Uint64("quota-inode", quota.INode).
		Msg("settings quota")

	err = setProjectQuota(q.backingFsBlockDev, projectID, quota)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't set project quota for target-path %s",
			targetPath)
		return
	}

	return
}

// setProjectQuota - set the quota for project id on xfs block device
func setProjectQuota(backingFsBlockDev string, projectID uint32, quota Quota) (err error) {
	var d C.fs_disk_quota_t

	d.d_version = C.FS_DQUOT_VERSION
	d.d_id = C.__u32(projectID)
	d.d_flags = C.XFS_PROJ_QUOTA

	d.d_fieldmask = C.FS_DQ_BHARD | C.FS_DQ_BSOFT | C.FS_DQ_ISOFT | C.FS_DQ_IHARD
	d.d_blk_hardlimit = C.__u64(quota.Size / 512)
	d.d_blk_softlimit = d.d_blk_hardlimit

	d.d_ino_hardlimit = C.__u64(quota.INode)
	d.d_ino_softlimit = d.d_ino_hardlimit

	var cs = C.CString(backingFsBlockDev)
	defer C.free(unsafe.Pointer(cs))

	_, _, errno := unix.Syscall6(unix.SYS_QUOTACTL, C.Q_XSETPQLIM,
		uintptr(unsafe.Pointer(cs)), uintptr(d.d_id),
		uintptr(unsafe.Pointer(&d)), 0, 0)
	if errno != 0 {
		err = errors.Errorf(
			"Failed to set quota limit for projid %d on %s: %v",
			projectID, backingFsBlockDev, errno.Error())
		return
	}

	return
}

// GetQuota - get the quota limits of a directory that was configured with SetQuota
func (q *Control) GetQuota(targetPath string, quota *Quota) (err error) {
	projectID, ok := q.quotas[targetPath]
	if !ok {
		err = errors.Errorf("projectId not found for path : %s", targetPath)
		return
	}

	//
	// get the quota limit for the container's project id
	//
	var d C.fs_disk_quota_t

	var cs = C.CString(q.backingFsBlockDev)
	defer C.free(unsafe.Pointer(cs))

	_, _, errno := unix.Syscall6(unix.SYS_QUOTACTL, C.Q_XGETPQUOTA,
		uintptr(unsafe.Pointer(cs)), uintptr(C.__u32(projectID)),
		uintptr(unsafe.Pointer(&d)), 0, 0)
	if errno != 0 {
		err = errors.Errorf(
			"Failed to get quota limit for projid %d on %s: %v",
			projectID, q.backingFsBlockDev, errno.Error())
		return
	}

	quota.Size = uint64(d.d_blk_hardlimit) * 512
	quota.INode = uint64(d.d_ino_hardlimit)
	return
}

// getProjectID - get the project id of path on xfs
func getProjectID(targetPath string) (uint32, error) {
	dir, err := openDir(targetPath)
	if err != nil {
		return 0, err
	}
	defer closeDir(dir)

	var fsx C.struct_fsxattr
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.FS_IOC_FSGETXATTR,
		uintptr(unsafe.Pointer(&fsx)))
	if errno != 0 {
		return 0, fmt.Errorf("Failed to get projid for %s: %v", targetPath, errno.Error())
	}

	return uint32(fsx.fsx_projid), nil
}

// setProjectID - set the project id of path on xfs
func setProjectID(targetPath string, projectID uint32) error {
	dir, err := openDir(targetPath)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var fsx C.struct_fsxattr
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.FS_IOC_FSGETXATTR,
		uintptr(unsafe.Pointer(&fsx)))
	if errno != 0 {
		return fmt.Errorf("Failed to get projid for %s: %v", targetPath, errno.Error())
	}
	fsx.fsx_projid = C.__u32(projectID)
	fsx.fsx_xflags |= C.FS_XFLAG_PROJINHERIT
	_, _, errno = unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.FS_IOC_FSSETXATTR,
		uintptr(unsafe.Pointer(&fsx)))
	if errno != 0 {
		return fmt.Errorf("Failed to set projid for %s: %v", targetPath, errno.Error())
	}

	return nil
}

// findNextProjectID - find the next project id to be used for containers
// by scanning driver home directory to find used project ids
func (q *Control) findNextProjectID(home string) error {
	files, err := ioutil.ReadDir(home)
	if err != nil {
		return fmt.Errorf("read directory failed : %s", home)
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		path := filepath.Join(home, file.Name())
		projid, err := getProjectID(path)
		if err != nil {
			return err
		}
		if projid > 0 {
			q.quotas[path] = projid
		}
		if q.nextProjectID <= projid {
			q.nextProjectID = projid + 1
		}
	}

	return nil
}

func free(p *C.char) {
	C.free(unsafe.Pointer(p))
}

func openDir(path string) (*C.DIR, error) {
	Cpath := C.CString(path)
	defer free(Cpath)

	dir := C.opendir(Cpath)
	if dir == nil {
		return nil, fmt.Errorf("Can't open dir")
	}
	return dir, nil
}

func closeDir(dir *C.DIR) {
	if dir != nil {
		C.closedir(dir)
	}
}

func getDirFd(dir *C.DIR) uintptr {
	return uintptr(C.dirfd(dir))
}

// makeBackingFsDev retrieves the backing block device of the
// driver home directory and creates a block device node under
// the home directory to be used by quotactl commands
func makeBackingFsDev(home string) (backingFsBlockDev string, err error) {
	fileinfo, err := os.Stat(home)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve information from file %s", home)
		return
	}

	backingFsBlockDev = path.Join(home, "backingFsBlockDev")

	// Re-create just in case someone copied the home directory over to a new device
	unix.Unlink(backingFsBlockDev)
	stat := fileinfo.Sys().(*syscall.Stat_t)

	err = unix.Mknod(backingFsBlockDev, unix.S_IFBLK|0600, int(stat.Dev))
	if err != nil {
		err = errors.Wrapf(err,
			"failed to mknod for block device %s",
			backingFsBlockDev)
		return
	}

	return
}
