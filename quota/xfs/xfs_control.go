// +build linux

// xfs implements XFS project quota controls for setting quota limits
// on a newly created directory.
package xfs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cirocosta/xfsvol/quota"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// blockDeviceName corresponds to the name of the
// special file that is meant to be used by xfs to
// keep track of the project quotas.
const blockDeviceName = "__control-device"

// Control gives the context to be used by storage driver
// who wants to apply project quotas to container dirs.
type Control struct {
	logger zerolog.Logger

	// backingFsBlockDev is the absolute path to the
	// block device that keeps track of quotas under
	// a given basePath (root of the project quota tree).
	backingFsBlockDev string

	// projectIdCache keeps track of the relation between
	// directories and project-ids.
	//
	// By making use of an in-memory cache we can avoid
	// using `cgo` too many times to just gather the
	// `projectId` of a given directory.
	projectIdCache map[string]uint32

	// lastProjectId keeps track of the last projectId
	// that has beem used. The purpose of it is to not
	// have a conflict with existing projectIds while at
	// the same time also have not to iterate over the
	// projectIdCache map.
	lastProjectId uint32
}

// ControlConfig specifies the configuration to be used by
// the controller that will hold the quota allocation state.
type ControlConfig struct {
	// StartingProjectId specifies the minimum projectid that
	// should be used in the projectid allocation.
	StartingProjectId *uint32

	// BasePath is the base in which all the directories
	// which quotas are applied get created from.
	//
	// Right in `BasePath` is also where a block device
	// is put to keep track of the quotas.
	BasePath string
}

// NewControl initializes project quota support under a given
// preconfigured BasePath.
//
// It does so by creating a block device right at BasePath
// and then having XFS manage quotas under this path by assigning
// project ids to each directory and binding such project ids
// with quotas.
func NewControl(cfg ControlConfig) (c Control, err error) {
	if cfg.BasePath == "" {
		err = errors.Errorf("BasePath must be provided")
		return
	}

	if cfg.StartingProjectId != nil {
		c.lastProjectId = *cfg.StartingProjectId
	}

	err = MakeBackingFsDev(cfg.BasePath, blockDeviceName)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create backingfs dev for base path %s",
			cfg.BasePath)
		return
	}

	c.backingFsBlockDev = filepath.Join(cfg.BasePath, blockDeviceName)
	c.logger = zerolog.New(os.Stdout).With().
		Str("from", "control").
		Logger()

	c.projectIdCache, err = GeneratePathToProjectIdMap(cfg.BasePath)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create projectid cache from basepath %s",
			cfg.BasePath)
		return
	}

	for _, projectId := range c.projectIdCache {
		if projectId > c.lastProjectId {
			c.lastProjectId = projectId
		}
	}

	c.logger.Debug().
		Str("base-path", cfg.BasePath).
		Uint32("last-project-id", c.lastProjectId).
		Msg("new control created")

	return
}

// GetBackingFsBlockDev retrieves the absolute path of the backing
// block device configured for the current quota control instance.
func (c *Control) GetBackingFsBlockDev() (blockDev string) {
	blockDev = c.backingFsBlockDev
	return
}

// GetQuota retrieves the quota settings associated with a targetPath
// that previously had a quota set for it.
//
// TODO differentiate between real errors and no quota being set
//	for the path.
func (c *Control) GetQuota(targetPath string) (q *quota.Quota, err error) {
	projectId, ok := c.projectIdCache[targetPath]
	if !ok {
		err = errors.Errorf(
			"no projectId associated with the path %s",
			targetPath)
		return
	}

	q, err = GetProjectQuota(c.backingFsBlockDev, projectId)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve quota")
		return
	}

	return
}

// SetQuota assigns a unique project id to a directory and then set the
// quota for that projectId.
func (c *Control) SetQuota(targetPath string, quota quota.Quota) (err error) {
	c.logger.Debug().Interface("cache", c.projectIdCache).Msg("will set quota")

	projectId, ok := c.projectIdCache[targetPath]
	if !ok {
		projectId = c.lastProjectId + 1
		err = SetProjectId(targetPath, projectId)
		if err != nil {
			err = errors.Wrapf(err,
				"couldn't set project id to path %s",
				targetPath)
			return
		}

		c.projectIdCache[targetPath] = projectId
		c.lastProjectId = projectId

		c.logger.Debug().Uint32("project-id", projectId).Msg("setting new project id")
	}

	c.logger.Debug().
		Uint32("project-id", projectId).
		Uint32("last-project-id", c.lastProjectId).
		Str("target-path", targetPath).
		Uint64("quota-size", quota.Size).
		Uint64("quota-inode", quota.INode).
		Msg("setting quota")

	err = SetProjectQuota(c.backingFsBlockDev, projectId, &quota)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't set project quota %+v for target-path %s",
			quota, targetPath)
		return
	}

	return
}

// GeneratePathToProjectIdMap creates a map that maps the
// projectIds associated with paths directly under a giving
// root path.
func GeneratePathToProjectIdMap(root string) (mapping map[string]uint32, err error) {
	var (
		files     []os.FileInfo
		absPath   string
		projectId uint32
	)

	files, err = ioutil.ReadDir(root)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to list files and directories under path %s",
			root)
		return
	}

	mapping = make(map[string]uint32)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		absPath = filepath.Join(root, file.Name())
		projectId, err = GetProjectId(absPath)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to retrieve projectid for directory %s",
				absPath)
			return
		}

		if projectId > 0 {
			mapping[absPath] = projectId
		}
	}

	return
}
