package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

// CreateFiles creates a bunch of files under a
// given base directory specified.
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

// WriteBytes writes a given number of bytes in the form
// of 'character'  to a given writer.
func WriteBytes(writer io.Writer, character byte, remaining uint64) (err error) {
	var chunkSize = uint64((1 << 12))
	var buf = make([]byte, chunkSize)
	var n = 0

	for ndx, _ := range buf {
		buf[ndx] = character
	}

	for remaining > 0 {
		if remaining < chunkSize {
			chunkSize = remaining
		}

		n, err = writer.Write(buf[:chunkSize])
		if err != nil {
			err = errors.Wrapf(err,
				"Couldn't write buf to writer - remaining %d",
				remaining)
			return
		}

		remaining -= uint64(n)
	}

	return
}
