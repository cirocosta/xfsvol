package manager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cirocosta/xfsvol/quota"
	"github.com/cirocosta/xfsvol/quota/xfs"
	"github.com/pkg/errors"
)

var (
	// NameRegex specifies a regular expression that is
	// used against every volume that is meant to be created
	// making sure that no weird names can be specified.
	NameRegex = regexp.MustCompile(`^[a-zA-Z0-9][\w\-]{1,250}$`)

	ErrInvalidName = errors.Errorf("Invalid name")
	ErrEmptyQuota  = errors.Errorf("Invalid quota")
	ErrEmptyINode  = errors.Errorf("Invalid inode")
	ErrNotFound    = errors.Errorf("Volume not found")
)

// Manager is the entity responsible for managing
// volumes under a given base path. It takes the
// responsability of 'CRUD'ing these volumes.
type Manager struct {
	quotaCtl quota.Control
	root     string
}

type FilesystemType uint8

const (
	UnknownFilesystemType FilesystemType = iota
	Ext4FilesystemType
	XfsFilesystemType
)

// Config represents the configuration to
// create a manager. It takes a root directory
// that is used to create a block device that
// controls project quotas bellow it.
type Config struct {
	Root string
	Type FilesystemType
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

// New instantiates a new manager that is meant to
// take care of volume creation, listing, updating
// and stats retrieval under a given root path.
func New(cfg Config) (manager Manager, err error) {
	if cfg.Root == "" {
		err = errors.Errorf("Root not specified.")
		return
	}

	if !filepath.IsAbs(cfg.Root) {
		err = errors.Errorf(
			"Root (%s) must be an absolute path",
			cfg.Root)
		return
	}

	if cfg.Type == Ext4FilesystemType {
		err = errors.Errorf("fs type ext4 is not implemented yet")
		return
	}

	quotaCtl, err := xfs.NewControl(xfs.ControlConfig{
		BasePath: cfg.Root,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't initialize quota control on root path %s",
			cfg.Root)
		return
	}

	manager.quotaCtl = &quotaCtl
	manager.root = cfg.Root

	return
}

// List lists all the volumes that have been created under
// a given root path that is controlled by this manager.
func (m Manager) List() (vols []Volume, err error) {
	files, err := ioutil.ReadDir(m.root)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't list files/directories from %s", m.root)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			var quota *quota.Quota
			absPath := filepath.Join(m.root, file.Name())

			quota, err = m.quotaCtl.GetQuota(absPath)
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

// Get tries to retrieve a volume by its name (not path).
//
// For instance, if the root path that the manager controls is
// `/mnt/xfs/volumes`, then a volume named `foo` that would live
// under `/mnt/xfs/volumes/foo` can be retrieved by passing `foo`
// as the parameter.
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
			var quota *quota.Quota
			absPath = filepath.Join(m.root, name)

			quota, err = m.quotaCtl.GetQuota(absPath)
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

// Create validates a volume specification and then proceed with
// creating the volume under the controlled root directory.
func (m Manager) Create(vol Volume) (absPath string, err error) {
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
		return
	}

	err = m.quotaCtl.SetQuota(absPath, quota.Quota{
		Size:  vol.Size,
		INode: vol.INode,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't set quota for volume name=%s size=%d inode=%d",
			vol.Name, vol.Size, vol.INode)
		os.RemoveAll(absPath)
		return
	}

	return
}

// Delete tries to delete a volume by its name.
//
// ps.: Deleting a volume that doesn't exist is considered
// an error.
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

// isValidName verifies whether a given name is considered valid
// or not based on a constant naming regular expression.
func isValidName(name string) bool {
	if name == "" {
		return false
	}

	return NameRegex.MatchString(name)
}
