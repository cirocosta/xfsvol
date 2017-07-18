package manager

import (
	"github.com/pkg/errors"

	units "github.com/docker/go-units"
)

func MustFromHumanSize(size string) uint64 {
	bytes, err := units.FromHumanSize(size)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't convert string in human size (size=%s) to bytes",
			size)
		panic(err)
	}

	return uint64(bytes)
}

func HumanSize(size uint64) string {
  return units.HumanSize(float64(size))
}
