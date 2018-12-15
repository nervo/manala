package syncer

import (
	"bytes"
	"errors"
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
)

var (
	Err            = errors.New("sync failed")
	ErrFileOverDir = errors.New("trying to overwrite a non-empty directory with a file")
)

type Interface interface {
	Sync(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) error
}

func New() Interface {
	return &syncer{
		delete:  true,
		noTimes: false,
	}
}

type syncer struct {
	// Set this to true to delete files in the destination that don't exist in the source.
	delete bool
	// By default, modification times are synced. This can be turned off by setting this to true.
	noTimes bool
}

func (snc *syncer) Sync(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) error {
	// Make sure src exists
	if _, err := srcFs.Stat(src); err != nil {
		return err
	}
	// Return error instead of replacing a non-empty directory with a file
	if b, err := snc.checkDir(dst, dstFs, src, srcFs); err != nil {
		return err
	} else if b {
		return ErrFileOverDir
	}

	return snc.sync(dst, dstFs, src, srcFs)
}

// Updates dst to match with src, handling both files and directories.
func (snc *syncer) sync(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) error {

	// Read files info
	dstat, err := dstFs.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	sstat, err := srcFs.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	if !sstat.IsDir() {
		// Src is a file
		// Delete dst if its a directory
		if dstat != nil && dstat.IsDir() {
			err = dstFs.RemoveAll(dst)
			if err != nil {
				return err
			}
		}

		eq, err := snc.equal(dst, dstFs, src, srcFs)
		if err != nil {
			return err
		}

		if !eq {
			// Create directory if needed.
			dir := filepath.Dir(dst)
			if dir != "." {
				err = dstFs.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
			}

			// Perform copy
			df, err := dstFs.Create(dst)
			if err != nil {
				return err
			}

			defer df.Close()

			sf, err := srcFs.Open(src)
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}

			defer sf.Close()

			_, err = io.Copy(df, sf)
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
		}

		return nil
	}

	// Src is a directory
	// Make dst if necessary
	if dstat == nil {
		// Dst does not exist; create directory
		err = dstFs.MkdirAll(dst, 0755)
		if err != nil {
			return err
		}
	} else if !dstat.IsDir() {
		// Dst is a file; remove and create directory
		err = dstFs.Remove(dst)
		if err != nil {
			return err
		}

		err = dstFs.MkdirAll(dst, 0755)
		if err != nil {
			return err
		}
	}

	// Go through sf files and sync them
	files, err := afero.ReadDir(srcFs, src)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// Make a map of filenames for quick lookup; used in deletion
	// Deletion below
	m := make(map[string]bool, len(files))
	for _, file := range files {
		dst2 := filepath.Join(dst, file.Name())
		src2 := filepath.Join(src, file.Name())
		if err = snc.sync(dst2, dstFs, src2, srcFs); err != nil {
			return err
		}
		m[file.Name()] = true
	}

	// Delete files from dst that does not exist in src
	if snc.delete {
		files, err = afero.ReadDir(dstFs, dst)
		if err != nil {
			return err
		}

		for _, file := range files {
			if !m[file.Name()] {
				err = dstFs.RemoveAll(filepath.Join(dst, file.Name()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Returns true if both files are equal
func (snc *syncer) equal(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) (bool, error) {
	// Get file infos
	dstInfo, dstErr := dstFs.Stat(dst)
	srcInfo, srcErr := srcFs.Stat(src)
	if os.IsNotExist(dstErr) || os.IsNotExist(srcErr) {
		return false, nil
	}
	if dstErr != nil {
		return false, dstErr
	}
	if srcErr != nil {
		return false, srcErr
	}

	// Check sizes
	if dstInfo.Size() != srcInfo.Size() {
		return false, nil
	}

	// Both have the same size, check the contents
	dstFile, err := dstFs.Open(dst)
	if err != nil {
		return false, err
	}

	defer dstFile.Close()

	srcFile, err := srcFs.Open(src)
	if err != nil {
		return false, err
	}

	defer srcFile.Close()

	dstBuf := make([]byte, 1000)
	srcBuf := make([]byte, 1000)

	for {
		// Read from both
		dstNb, err := dstFile.Read(dstBuf)
		if err != nil && err != io.EOF {
			return false, err
		}

		srcNb, err := srcFile.Read(srcBuf)
		if err != nil && err != io.EOF {
			return false, err
		}

		// Compare read bytes
		if !bytes.Equal(dstBuf[:dstNb], srcBuf[:srcNb]) {
			return false, nil
		}

		// End of both files
		if dstNb == 0 && srcNb == 0 {
			break
		}
	}

	return true, nil
}

// Returns true if dst is a non-empty directory and src is a file
func (snc *syncer) checkDir(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) (bool, error) {
	// Read file info
	dstat, err := dstFs.Stat(dst)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	sstat, err := srcFs.Stat(src)
	if err != nil {
		return false, err
	}

	// Return false is dst is not a directory or src is a directory
	if !dstat.IsDir() || sstat.IsDir() {
		return false, nil
	}

	// Dst is a directory and src is a file
	// Check if dst is non-empty
	// Read dst directory
	files, err := afero.ReadDir(dstFs, dst)
	if err != nil {
		return false, err
	}
	if len(files) > 0 {
		return true, nil
	}

	return false, nil
}
