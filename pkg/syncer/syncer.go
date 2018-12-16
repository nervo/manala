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
	Err = errors.New("sync failed")
)

type SourceNotExistError struct {
	Source string
}

func (e *SourceNotExistError) Error() string {
	return "no source " + e.Source + " file or directory "
}

type SourceFileOverDestinationDirectoryError struct {
	Source      string
	Destination string
}

func (e *SourceFileOverDestinationDirectoryError) Error() string {
	return "source " + e.Source + " file over " + e.Destination + " directory "
}

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
	return snc.sync(dst, dstFs, src, srcFs)
}

// Updates dst to match with src, handling both files and directories.
func (snc *syncer) sync(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) error {
	// Source info
	srcInfo, err := srcFs.Stat(src)

	if err != nil {
		// Source does not exist
		if os.IsNotExist(err) {
			return &SourceNotExistError{src}
		} else {
			return err
		}
	}

	// Destination info
	dstInfo, err := dstFs.Stat(dst)

	// Error other than not existing destination
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Source is a file
	if !srcInfo.IsDir() {
		// Delete destination if it's a non empty directory
		if dstInfo != nil && dstInfo.IsDir() {
			files, err := afero.ReadDir(dstFs, dst)
			if err != nil {
				return err
			}
			if len(files) > 0 {
				return &SourceFileOverDestinationDirectoryError{src, dst}
			}
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
			dstDir := filepath.Dir(dst)
			if dstDir != "." {
				err = dstFs.MkdirAll(dstDir, 0755)
				if err != nil {
					return err
				}
			}

			// Perform copy
			dstFile, err := dstFs.Create(dst)
			if err != nil {
				return err
			}

			defer dstFile.Close()

			srcFile, err := srcFs.Open(src)
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}

			defer srcFile.Close()

			_, err = io.Copy(dstFile, srcFile)
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
	if dstInfo == nil {
		// Dst does not exist; create directory
		err = dstFs.MkdirAll(dst, 0755)
		if err != nil {
			return err
		}
	} else if !dstInfo.IsDir() {
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
	// Info
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
	// Todo: switch to checksum (https://golang.org/pkg/crypto/md5/#example_New_file)
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
