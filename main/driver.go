package main

import (
	"fmt"
	"sync"

	"github.com/cirocosta/xfsvol/manager"
	"github.com/pkg/errors"
	"github.com/ventu-io/go-shortid"

	v "github.com/docker/go-plugins-helpers/volume"
	log "github.com/sirupsen/logrus"
)

const (
	HostMountPoint = "/mnt/nfs"
)

type nfsVolDriver struct {
	logger  *log.Entry
	manager *manager.Manager
	sync.Mutex
}

func newNfsVolDriver() (d nfsVolDriver, err error) {
	m, err := manager.New(manager.Config{
		Root: HostMountPoint,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't initiate fs manager mounting at %s",
			HostMountPoint)
		return
	}

	d.logger = log.WithField("from", "driver")
	d.logger.Info("driver initiated")
	d.manager = &m

	return
}

func (d nfsVolDriver) Create(req v.Request) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "create").
		WithField("name", req.Name).
		WithField("opts", req.Options)
	logger.Debug("start")

	d.Lock()
	defer d.Unlock()

	abs, err := d.manager.Create(req.Name)
	if err != nil {
		logger.WithError(err).Error("couldn't create volume")
		resp.Err = err.Error()
		return
	}

	logger.
		WithField("abs", abs).
		Debug("finish")
	return
}

func (d nfsVolDriver) List(req v.Request) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "list").
		WithField("name", req.Name).
		WithField("opts", req.Options)
	logger.Debug("start")

	d.Lock()
	defer d.Unlock()

	dirs, err := d.manager.List()
	if err != nil {
		logger.WithError(err).Error("couldn't list volumes")
		resp.Err = err.Error()
		return
	}

	resp.Volumes = make([]*v.Volume, len(dirs))
	for idx, dir := range dirs {
		resp.Volumes[idx] = &v.Volume{
			Name: dir,
		}
	}

	logger.
		WithField("number-of-volumes", len(dirs)).
		Debug("finish")
	return
}

func (d nfsVolDriver) Get(req v.Request) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "get").
		WithField("name", req.Name).
		WithField("opts", req.Options)
	logger.Debug("start")

	d.Lock()
	defer d.Unlock()

	mp, found, err := d.manager.Get(req.Name)
	if err != nil {
		logger.WithError(err).Error("errored retrieving path for volume")
		resp.Err = err.Error()
		return
	}

	if !found {
		logger.WithError(err).Info("volume not found")
		resp.Err = fmt.Sprintf("volume %s not found", req.Name)
		return
	}

	resp.Volume = &v.Volume{
		Name:       req.Name,
		Mountpoint: mp,
	}

	logger.
		WithField("mountpoint", mp).
		Debug("finish")
	return
}

func (d nfsVolDriver) Remove(req v.Request) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "remove").
		WithField("name", req.Name).
		WithField("opts", req.Options)
	logger.Debug("start")

	d.Lock()
	defer d.Unlock()

	err := d.manager.Delete(req.Name)
	if err != nil {
		logger.WithError(err).Error("errored trying to delete volume")
		resp.Err = err.Error()
		return
	}

	logger.Debug("finish")
	return
}

func (d nfsVolDriver) Path(req v.Request) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "path").
		WithField("name", req.Name).
		WithField("method", "path")
	logger.Debug("start")

	d.Lock()
	defer d.Unlock()

	mp, found, err := d.manager.Get(req.Name)
	if err != nil {
		logger.WithError(err).Error("errored retrieving path for volume")
		resp.Err = err.Error()
		return
	}

	if !found {
		logger.WithError(err).Info("volume not found")
		resp.Err = fmt.Sprintf("volume %s not found", req.Name)
		return
	}

	logger.Debug("finish")

	resp.Mountpoint = mp
	return
}

func (d nfsVolDriver) Mount(req v.MountRequest) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "mount").
		WithField("name", req.Name).
		WithField("id", req.ID)
	logger.Debug("start")

	d.Lock()
	defer d.Unlock()

	mp, found, err := d.manager.Get(req.Name)
	if err != nil {
		logger.WithError(err).Error("errored retrieving path for volume")
		resp.Err = err.Error()
		return
	}

	if !found {
		logger.WithError(err).Info("volume not found")
		resp.Err = fmt.Sprintf("volume %s not found", req.Name)
		return
	}

	logger.Debug("finish")
	resp.Mountpoint = mp
	return
}

func (d nfsVolDriver) Unmount(req v.UnmountRequest) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "unmount").
		WithField("name", req.Name).
		WithField("id", req.ID)
	logger.Debug("start")
	logger.Debug("finish")

	d.Lock()
	defer d.Unlock()

	return
}

func (d nfsVolDriver) Capabilities(req v.Request) (resp v.Response) {
	var logger = d.logger.
		WithField("log-id", shortid.MustGenerate()).
		WithField("method", "capabilities").
		WithField("name", req.Name)
	logger.Debug("start")

	resp.Capabilities = v.Capability{
		Scope: "global",
	}
	return
}
