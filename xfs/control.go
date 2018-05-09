// +build linux

// xfs implements XFS project quota controls for setting quota limits
// on a newly created directory.
package xfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// blockDeviceName corresponds to the name of the
// special file that is meant to be used by xfs to
// keep track of the project quotas.
const blockDeviceName = "backingFsBlockDev"

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
func NewControl(cfg ControlConfig) (c Control, err error) {
	var minProjectID uint32

	if cfg.BasePath == "" {
		err = errors.Errorf("BasePath must be provided")
		return
	}

	c.logger = zerolog.New(os.Stdout).With().Str("from", "control").Logger()

	if cfg.StartingProjectId == nil {
		minProjectID, err = GetProjectId(cfg.BasePath)
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
	err = MakeBackingFsDev(cfg.BasePath, blockDeviceName)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create backingfs dev for base path %s",
			cfg.BasePath)
		return
	}

	c.backingFsBlockDev = filepath.Join(cfg.BasePath, blockDeviceName)

	//
	// Test if filesystem supports project quotas by trying to set
	// a quota on the first available project id
	err = SetProjectQuota(c.backingFsBlockDev, minProjectID, &Quota{
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

func (c *Control) GetQuota(targetPath string) (q *Quota, err error) {
	projectID, ok := c.quotas[targetPath]
	if !ok {
		err = errors.Errorf("projectId not found for path : %s", targetPath)
		return
	}

	q, err = GetProjectQuota(c.backingFsBlockDev, projectID)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve quota")
		return
	}

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
		err = SetProjectId(targetPath, projectID)
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

	err = SetProjectQuota(q.backingFsBlockDev, projectID, &quota)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't set project quota for target-path %s",
			targetPath)
		return
	}

	return
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
		projid, err := GetProjectId(path)
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
