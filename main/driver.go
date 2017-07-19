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
	HostMountPoint = "/mnt/xfs/volumes"
	DefaultSize    = "1GiB"
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

	size, present := req.Options["size"]
	if !present {
		size = DefaultSize
	}

	sizeInBytes, err := manager.FromHumanSize(size)
	if err != nil {
		err = errors.Errorf(
			"couldn't convert specified size [%s] into bytes",
			size)
		logger.WithError(err).Error("couldn't create volume")
		resp.Err = err.Error()
		return
	}

	d.Lock()
	defer d.Unlock()

	abs, err := d.manager.Create(manager.Volume{
		Name: req.Name,
		Size: sizeInBytes,
	})
	if err != nil {
		logger.WithError(err).Error("couldn't create volume")
		resp.Err = err.Error()
		return
	}

	logger.
		WithField("abs", abs).
		Debug("finished creating volume")
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

	vols, err := d.manager.List()
	if err != nil {
		logger.WithError(err).Error("couldn't list volumes")
		resp.Err = err.Error()
		return
	}

	resp.Volumes = make([]*v.Volume, len(vols))
	for idx, vol := range vols {
		resp.Volumes[idx] = &v.Volume{
			Name: vol.Name,
		}
	}

	logger.
		WithField("number-of-volumes", len(vols)).
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

	vol, found, err := d.manager.Get(req.Name)
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
		Mountpoint: vol.Path,
	}

	logger.
		WithField("mountpoint", vol.Path).
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

	vol, found, err := d.manager.Get(req.Name)
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

	resp.Mountpoint = vol.Path
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

	vol, found, err := d.manager.Get(req.Name)
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
	resp.Mountpoint = vol.Path
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
