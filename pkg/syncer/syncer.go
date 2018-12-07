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
	Sync(path string, dst afero.Fs, src afero.Fs) error
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

func (s *syncer) Sync(path string, dst afero.Fs, src afero.Fs) error {
	// Make sure src path exists
	if _, err := src.Stat(path); err != nil {
		return err
	}
	// Return error instead of replacing a non-empty directory with a file
	if b, err := s.checkDir(path, dst, src); err != nil {
		return err
	} else if b {
		return ErrFileOverDir
	}

	return s.sync(path, dst, src)
}

// Updates dst to match with src, handling both files and directories.
func (s *syncer) sync(path string, dst afero.Fs, src afero.Fs) error {

	// Read files info
	dstat, err := dst.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	sstat, err := src.Stat(path)
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
			err = dst.RemoveAll(path)
			if err != nil {
				return err
			}
		}

		eq, err := s.equal(path, dst, src)
		if err != nil {
			return err
		}

		if !eq {
			// Create directory if needed.
			dir := filepath.Dir(path)
			if dir != "." {
				err = dst.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
			}

			// Perform copy
			df, err := dst.Create(path)
			if err != nil {
				return err
			}

			defer df.Close()

			sf, err := src.Open(path)
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
		err = dst.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	} else if !dstat.IsDir() {
		// Dst is a file; remove and create directory
		err = dst.Remove(path)
		if err != nil {
			return err
		}

		err = dst.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	// Go through sf files and sync them
	files, err := afero.ReadDir(src, path)
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
		path2 := filepath.Join(path, file.Name())
		if err = s.sync(path2, dst, src); err != nil {
			return err
		}
		m[file.Name()] = true
	}

	// Delete files from dst that does not exist in src
	if s.delete {
		files, err = afero.ReadDir(dst, path)
		if err != nil {
			return err
		}

		for _, file := range files {
			if !m[file.Name()] {
				err = dst.RemoveAll(filepath.Join(path, file.Name()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Returns true if both files are equal
func (s *syncer) equal(path string, a afero.Fs, b afero.Fs) (bool, error) {
	// Get file infos
	info1, err1 := a.Stat(path)
	info2, err2 := b.Stat(path)
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		return false, nil
	}
	if err1 != nil {
		return false, err1
	}
	if err2 != nil {
		return false, err2
	}

	// Check sizes
	if info1.Size() != info2.Size() {
		return false, nil
	}

	// Both have the same size, check the contents
	f1, err := a.Open(path)
	if err != nil {
		return false, err
	}

	defer f1.Close()

	f2, err := b.Open(path)
	if err != nil {
		return false, err
	}

	defer f2.Close()

	buf1 := make([]byte, 1000)
	buf2 := make([]byte, 1000)

	for {
		// Read from both
		n1, err := f1.Read(buf1)
		if err != nil && err != io.EOF {
			return false, err
		}

		n2, err := f2.Read(buf2)
		if err != nil && err != io.EOF {
			return false, err
		}

		// Compare read bytes
		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		// End of both files
		if n1 == 0 && n2 == 0 {
			break
		}
	}

	return true, nil
}

// Returns true if dst is a non-empty directory and src is a file
func (s *syncer) checkDir(path string, dst afero.Fs, src afero.Fs) (bool, error) {
	// Read file info
	dstat, err := dst.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	sstat, err := src.Stat(path)
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
	files, err := afero.ReadDir(dst, path)
	if err != nil {
		return false, err
	}
	if len(files) > 0 {
		return true, nil
	}

	return false, nil
}
