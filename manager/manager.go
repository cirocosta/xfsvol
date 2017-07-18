package manager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cirocosta/xfsvol/lib"
	"github.com/pkg/errors"
	_ "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type Manager struct {
	quotaCtl *lib.Control
	root     string
}

type Config struct {
	Root string
}

type Volume struct {
	Name string
	Size uint64
}

var (
	NameRegex = regexp.MustCompile(`^[a-zA-Z][\w\-]{1,30}$`)

	ErrInvalidName = errors.Errorf("Invalid name")
	ErrEmptyQuota  = errors.Errorf("Invalid quota - Can't be 0")
	ErrNotFound    = errors.Errorf("Volume not found")
)

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

	err = unix.Access(cfg.Root, unix.W_OK)
	if err != nil {
		err = errors.Wrapf(err,
			"Root (%s) must be writable.",
			cfg.Root)
		return
	}

	quotaCtl, err := lib.NewControl(cfg.Root)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't initialize XFS quota control on root path %s",
			cfg.Root)
		return
	}

	manager.quotaCtl = quotaCtl
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
			quota := lib.Quota{}
			absPath := filepath.Join(m.root, file.Name())

			err = m.quotaCtl.GetQuota(absPath, &quota)
			if err != nil {
				err = errors.Wrapf(err,
					"Couldn't retrieve quota for directory %s",
					file.Name())
				return
			}

			vols = append(vols, Volume{
				Name: file.Name(),
				Size: quota.Size,
			})
		}
	}

	return
}

func (m Manager) Get(name string) (absPath string, found bool, err error) {
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
			absPath = filepath.Join(m.root, name)
			return
		}
	}

	return
}

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

	err = m.quotaCtl.SetQuota(absPath, lib.Quota{
		Size: vol.Size,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't set quota for volume name=%s size=%d",
			vol.Name, vol.Size)
		os.RemoveAll(absPath)
		return
	}

	return
}

func (m Manager) Delete(name string) (err error) {
	if !isValidName(name) {
		err = ErrInvalidName
		return
	}

	abs, found, err := m.Get(name)
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

	err = os.RemoveAll(abs)
	if err != nil {
		err = errors.Wrapf(err,
			"Errored removing volume named %s at path %s",
			name, abs)
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
