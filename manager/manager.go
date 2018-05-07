package manager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cirocosta/xfsvol/xfs"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"
)

var (
	NameRegex = regexp.MustCompile(`^[a-zA-Z0-9][\w\-]{1,250}$`)

	ErrInvalidName = errors.Errorf("Invalid name")
	ErrEmptyQuota  = errors.Errorf("Invalid quota - Can't be 0")
	ErrEmptyINode  = errors.Errorf("Invalid inode - Can't be 0")
	ErrNotFound    = errors.Errorf("Volume not found")
)

// Manager is the entity responsible for managing
// volumes under a given base path. It takes the
// responsability of 'CRUD'ing these volumes.
type Manager struct {
	logger   *log.Entry
	quotaCtl *xfs.Control
	root     string
}

// Config represents the configuration to
// create a manager. It takes a root directory
// that is used to create a block device that
// controls project quotas bellow it.
type Config struct {
	Root string
}

// Volume represents a volume under a given
// root path that has a project quota associated
// with it.
type Volume struct {
	Name  string
	Path  string
	Size  uint64
	INode uint64
}

func New(cfg Config) (manager Manager, err error) {
	if cfg.Root == "" {
		err = errors.Errorf("Root not specified.")
		return
	}

	manager.logger = log.WithField("root", cfg.Root)

	if !filepath.IsAbs(cfg.Root) {
		err = errors.Errorf(
			"Root (%s) must be an absolute path",
			cfg.Root)
		return
	}

	err = unix.Access(cfg.Root, unix.W_OK)
	if err != nil {
		err = errors.Wrapf(err,
			"Root (%s) must be writable.",
			cfg.Root)
		return
	}

	quotaCtl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath: cfg.Root,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't initialize XFS quota control on root path %s",
			cfg.Root)
		return
	}

	manager.logger.Debug("manager initialized")
	manager.quotaCtl = &quotaCtl
	manager.root = cfg.Root
	return
}

func (m Manager) List() (vols []Volume, err error) {
	files, err := ioutil.ReadDir(m.root)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't list files/directories from %s", m.root)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			quota := xfs.Quota{}
			absPath := filepath.Join(m.root, file.Name())

			err = m.quotaCtl.GetQuota(absPath, &quota)
			if err != nil {
				err = errors.Wrapf(err,
					"Couldn't retrieve quota for directory %s",
					file.Name())
				return
			}

			vols = append(vols, Volume{
				Name:  file.Name(),
				Size:  quota.Size,
				INode: quota.INode,
				Path:  absPath,
			})
		}
	}

	return
}

func (m Manager) Get(name string) (vol Volume, found bool, err error) {
	var absPath string

	if !isValidName(name) {
		err = ErrInvalidName
		return
	}

	files, err := ioutil.ReadDir(m.root)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't list files/directories from %s", m.root)
		return
	}

	for _, file := range files {
		if file.IsDir() && file.Name() == name {
			found = true
			quota := xfs.Quota{}
			absPath = filepath.Join(m.root, name)

			err = m.quotaCtl.GetQuota(absPath, &quota)
			if err != nil {
				err = errors.Wrapf(err,
					"Couldn't retrieve quota for directory %s",
					file.Name())
				return
			}

			vol.Name = filepath.Base(absPath)
			vol.Size = quota.Size
			vol.INode = quota.INode
			vol.Path = absPath
			return
		}
	}

	return
}

func (m Manager) Create(vol Volume) (absPath string, err error) {
	var log = m.logger.
		WithField("name", vol.Name).
		WithField("size", vol.Size).
		WithField("inode", vol.INode)

	log.Debug("starting volume creation")

	if vol.Size == 0 {
		err = ErrEmptyQuota
		return
	}

	if !isValidName(vol.Name) {
		err = ErrInvalidName
		return
	}

	absPath = filepath.Join(m.root, vol.Name)
	err = os.MkdirAll(absPath, 0755)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't create directory %s", absPath)
		log.WithError(err).Error("volume creation failed")
		return
	}

	err = m.quotaCtl.SetQuota(absPath, xfs.Quota{
		Size:  vol.Size,
		INode: vol.INode,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't set quota for volume name=%s size=%d inode=%d",
			vol.Name, vol.Size, vol.INode)
		log.WithError(err).Error("volume creation failed")
		os.RemoveAll(absPath)
		return
	}

	log.Debug("volume created")
	return
}

func (m Manager) Delete(name string) (err error) {
	if !isValidName(name) {
		err = ErrInvalidName
		return
	}

	vol, found, err := m.Get(name)
	if err != nil {
		err = errors.Wrapf(err,
			"Errored retrieving abs path for name %s",
			name)
		return
	}

	if !found {
		err = ErrNotFound
		return
	}

	err = os.RemoveAll(vol.Path)
	if err != nil {
		err = errors.Wrapf(err,
			"Errored removing volume named %s at path %s",
			name, vol.Path)
		return
	}

	return
}

func isValidName(name string) bool {
	if name == "" {
		return false
	}

	return NameRegex.MatchString(name)
}
