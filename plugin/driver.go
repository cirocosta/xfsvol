package main

import (
	"os"
	"sync"

	"github.com/cirocosta/xfsvol/manager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/ventu-io/go-shortid"

	v "github.com/docker/go-plugins-helpers/volume"
)

type DriverConfig struct {
	HostMountpoint string
	DefaultSize    string
}

type Driver struct {
	defaultSize string
	logger      zerolog.Logger
	manager     *manager.Manager
	sync.Mutex
}

func NewDriver(cfg DriverConfig) (d Driver, err error) {
	if cfg.HostMountpoint == "" {
		err = errors.Errorf("HostMountpoint must be specified")
		return
	}

	if cfg.DefaultSize == "" {
		err = errors.Errorf("DefaultSize must be specified")
		return
	}

	m, err := manager.New(manager.Config{
		Root: cfg.HostMountpoint,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't initiate fs manager mounting at %s",
			cfg.HostMountpoint)
		return
	}

	d.logger = zerolog.New(os.Stdout).With().Str("from", "driver").Logger()
	d.defaultSize = cfg.DefaultSize
	d.logger.Info().Msg("driver initiated")
	d.manager = &m

	return
}

func (d Driver) Create(req *v.CreateRequest) (err error) {
	var logger = d.logger.With().
		Str("log-id", shortid.MustGenerate()).
		Str("method", "create").
		Str("name", req.Name).
		Str("opts-size", req.Options["size"]).
		Str("opts-inode", req.Options["inode"]).
		Logger()

	size, present := req.Options["size"]
	if !present {
		logger.Debug().
			Str("default", d.defaultSize).
			Msg("no size opt found, using default")
		size = d.defaultSize
	}

	sizeInBytes, err := manager.FromHumanSize(size)
	if err != nil {
		err = errors.Errorf(
			"couldn't convert specified size [%s] into bytes",
			size)
		return
	}

	d.Lock()
	defer d.Unlock()

	logger.Debug().
		Msg("starting creation")

	absHostPath, err := d.manager.Create(manager.Volume{
		Name: req.Name,
		Size: sizeInBytes,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to create volume %s",
			req.Name)
		return
	}

	logger.Debug().
		Str("abs-host-path", absHostPath).
		Msg("finished creating volume")
	return
}

func (d Driver) List() (resp *v.ListResponse, err error) {
	var logger = d.logger.With().
		Str("log-id", shortid.MustGenerate()).
		Str("method", "list").
		Logger()

	d.Lock()
	defer d.Unlock()

	logger.Debug().
		Msg("starting volume listing")

	vols, err := d.manager.List()
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to list volumes")
		return
	}

	resp = new(v.ListResponse)
	resp.Volumes = make([]*v.Volume, len(vols))
	for idx, vol := range vols {
		resp.Volumes[idx] = &v.Volume{
			Name: vol.Name,
		}
	}

	logger.Debug().
		Int("number-of-volumes", len(vols)).
		Msg("finished listing volumes")
	return
}

func (d Driver) Get(req *v.GetRequest) (resp *v.GetResponse, err error) {
	var logger = d.logger.With().
		Str("log-id", shortid.MustGenerate()).
		Str("method", "get").
		Str("name", req.Name).
		Logger()

	d.Lock()
	defer d.Unlock()

	logger.Debug().
		Msg("starting volume retrieval")

	vol, found, err := d.manager.Get(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to get volume named %s",
			req.Name)
		return
	}

	if !found {
		err = errors.Errorf("volume %s not found", req.Name)
		return
	}

	resp = new(v.GetResponse)
	resp.Volume = &v.Volume{
		Name:       req.Name,
		Mountpoint: vol.Path,
	}

	logger.Debug().
		Str("mountpoint", vol.Path).
		Msg("finished retrieving volume")
	return
}

func (d Driver) Remove(req *v.RemoveRequest) (err error) {
	var logger = d.logger.With().
		Str("log-id", shortid.MustGenerate()).
		Str("method", "remove").
		Str("name", req.Name).
		Logger()

	d.Lock()
	defer d.Unlock()

	logger.Debug().
		Msg("starting removal")

	err = d.manager.Delete(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to delete volume named %s",
			req.Name)
		return
	}

	logger.Debug().
		Msg("finished removing volume")
	return
}

func (d Driver) Path(req *v.PathRequest) (resp *v.PathResponse, err error) {
	var logger = d.logger.With().
		Str("log-id", shortid.MustGenerate()).
		Str("method", "path").
		Str("name", req.Name).
		Logger()

	d.Lock()
	defer d.Unlock()

	logger.Debug().
		Msg("starting path retrieval")

	vol, found, err := d.manager.Get(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to retrieve volume named %s",
			req.Name)
		return
	}

	if !found {
		err = errors.Errorf("volume %s not found", req.Name)
		return
	}

	logger.Debug().
		Str("path", vol.Path).
		Msg("finished retrieving volume path")

	resp = new(v.PathResponse)
	resp.Mountpoint = vol.Path
	return
}

func (d Driver) Mount(req *v.MountRequest) (resp *v.MountResponse, err error) {
	var logger = d.logger.With().
		Str("log-id", shortid.MustGenerate()).
		Str("method", "mount").
		Str("name", req.Name).
		Str("id", req.ID).
		Logger()

	d.Lock()
	defer d.Unlock()

	logger.Debug().
		Msg("starting mount")

	vol, found, err := d.manager.Get(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve volume named %s",
			req.Name)
		return
	}

	if !found {
		err = errors.Errorf("volume %s not found", req.Name)
		return
	}

	logger.Debug().
		Str("mountpoint", vol.Path).
		Msg("finished mounting volume")

	resp = new(v.MountResponse)
	resp.Mountpoint = vol.Path
	return
}

func (d Driver) Unmount(req *v.UnmountRequest) (err error) {
	var logger = d.logger.With().
		Str("log-id", shortid.MustGenerate()).
		Str("method", "mount").
		Str("name", req.Name).
		Str("id", req.ID).
		Logger()

	d.Lock()
	defer d.Unlock()

	logger.Debug().Msg("started unmounting")
	logger.Debug().Msg("finished unmounting")

	return
}

// TODO is it global?
func (d Driver) Capabilities() (resp *v.CapabilitiesResponse) {
	resp = &v.CapabilitiesResponse{
		Capabilities: v.Capability{
			Scope: "global",
		},
	}
	return
}
