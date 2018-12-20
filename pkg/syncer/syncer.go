package syncer

import (
	"bytes"
	"crypto/md5"
	"errors"
	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"github.com/spf13/afero"
	"html/template"
	"os"
	"path/filepath"
	"strings"
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

type FileHookFunc func(src string, srcData []byte, dst string) (string, []byte, string, error)

type Interface interface {
	Sync(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) error
	SetFileHook(hook FileHookFunc)
	TemplateHook(data interface{}) FileHookFunc
}

func New(logger log.Interface) Interface {
	return &syncer{
		delete: true,
		logger: logger,
	}
}

type syncer struct {
	// Set this to true to delete files in the destination that don't exist in the source.
	delete bool
	// File hook
	fileHook FileHookFunc
	// Logger
	logger log.Interface
}

func (snc *syncer) SetFileHook(hook FileHookFunc) {
	snc.fileHook = hook
}

// Updates dst to match with src, handling both files and directories.
func (snc *syncer) Sync(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) error {
	// Source info
	srcInfo, srcErr := srcFs.Stat(src)

	if srcErr != nil {
		// Source does not exist
		if os.IsNotExist(srcErr) {
			return &SourceNotExistError{src}
		} else {
			return srcErr
		}
	}

	/* ********* */
	/* Directory */
	/* ********* */

	if srcInfo.IsDir() {
		snc.logger.WithFields(log.Fields{
			"src": src,
			"dst": dst,
		}).Info("Sync directory")

		// Destination info
		dstInfo, dstErr := dstFs.Stat(dst)

		// Error other than not existing destination
		if dstErr != nil && !os.IsNotExist(dstErr) {
			return dstErr
		}

		// Make destination if necessary
		if dstInfo == nil {
			// Destination does not exist; create directory
			err := dstFs.MkdirAll(dst, 0755)
			if err != nil {
				return err
			}
		} else if !dstInfo.IsDir() {
			// Destination is a file; remove and create directory
			err := dstFs.Remove(dst)
			if err != nil {
				return err
			}

			err = dstFs.MkdirAll(dst, 0755)
			if err != nil {
				return err
			}
		}

		// Go through source files and sync them
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
			dstFile := filepath.Join(dst, file.Name())
			srcFile := filepath.Join(src, file.Name())
			dstFile, err = snc.syncFile(dstFile, dstFs, srcFile, srcFs)
			if err != nil {
				return err
			}
			m[filepath.Base(dstFile)] = true
		}

		// Delete files from destination that does not exist in source
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

	/* **** */
	/* File */
	/* **** */

	_, err := snc.syncFile(dst, dstFs, src, srcFs)

	return err
}

func (snc *syncer) syncFile(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) (string, error) {
	snc.logger.WithFields(log.Fields{
		"src": src,
		"dst": dst,
	}).Info("Sync file")

	// Data
	srcData, err := afero.ReadFile(srcFs, src)
	if err != nil {
		return "", err
	}

	// File hook
	if snc.fileHook != nil {
		src, srcData, dst, err = snc.fileHook(src, srcData, dst)
		if err != nil {
			return "", err
		}
	}

	// Destination info
	dstInfo, dstErr := dstFs.Stat(dst)

	// Error other than not existing destination
	if dstErr != nil && !os.IsNotExist(dstErr) {
		return "", dstErr
	}

	// Delete destination if it's a directory
	if dstInfo != nil && dstInfo.IsDir() {
		err = dstFs.RemoveAll(dst)
		if err != nil {
			return "", err
		}

		// Destination info
		dstInfo, dstErr = dstFs.Stat(dst)

		// Error other than not existing destination
		if dstErr != nil && !os.IsNotExist(dstErr) {
			return "", dstErr
		}
	}

	eq, err := snc.equal(dst, dstFs, dstInfo, dstErr, srcData)
	if err != nil {
		return "", err
	}

	if !eq {
		// Create directory if needed.
		dstDir := filepath.Dir(dst)
		if dstDir != "." {
			err = dstFs.MkdirAll(dstDir, 0755)
			if err != nil {
				return "", err
			}
		}

		err := afero.WriteFile(dstFs, dst, srcData, 0666)
		if err != nil {
			return "", err
		}
	}

	return dst, nil
}

func (snc *syncer) equal(dst string, dstFs afero.Fs, dstInfo os.FileInfo, dstErr error, srcData []byte) (bool, error) {
	// Destination does not exists
	if os.IsNotExist(dstErr) {
		return false, nil
	}
	// Source data and destination file size differs
	if int(dstInfo.Size()) != len(srcData) {
		return false, nil
	}
	// Checksums differs
	dstData, err := afero.ReadFile(dstFs, dst)
	if err != nil {
		return false, err
	}
	dstHash := md5.Sum(dstData)
	srcHash := md5.Sum(srcData)

	if dstHash != srcHash {
		return false, nil
	}

	return true, nil
}

func (snc *syncer) TemplateHook(data interface{}) FileHookFunc {
	return func(src string, srcData []byte, dst string) (string, []byte, string, error) {
		// Filter on ".tmpl" source files
		if filepath.Ext(src) != ".tmpl" {
			return src, srcData, dst, nil
		}

		// Remove destination ".tmpl" extension
		dst = strings.TrimSuffix(dst, ".tmpl")

		snc.logger.WithFields(log.Fields{
			"src": src,
			"dst": dst,
		}).Info("Sync file template")

		tmpl, err := template.New(src).Funcs(sprig.FuncMap()).Parse(string(srcData))
		if err != nil {
			return "", nil, "", err
		}

		var tmplData bytes.Buffer

		err = tmpl.Execute(&tmplData, data)
		if err != nil {
			return "", nil, "", err
		}

		srcData = tmplData.Bytes()

		return src, srcData, dst, nil
	}
}
