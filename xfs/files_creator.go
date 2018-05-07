package xfs

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func CreateFiles(base string, n int) (err error) {
	var (
		file  *os.File
		fname string
	)

	for n > 0 {
		fname = fmt.Sprintf("%s/%d", base, n)

		file, err = os.Create(fname)
		if err != nil {
			err = errors.Wrapf(err, "couldn't create file %s", fname)
			return
		}
		file.Close()
		n--
	}

	return
}
