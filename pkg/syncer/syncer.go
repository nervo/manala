package syncer

import (
	"bytes"
	"crypto/md5"
	"errors"
	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"manala/pkg/project"
	"manala/pkg/template"
	"os"
	"path/filepath"
	"strings"
	engine "text/template"
)

/**********/
/* Errors */
/**********/

var (
	Err = errors.New("sync failed")
)

type SourceNotExistError struct {
	Source string
}

func (e *SourceNotExistError) Error() string {
	return "no source " + e.Source + " file or directory "
}

/*********/
/* Hooks */
/*********/

type FileHookFunc func(src string, srcContent []byte, dst string) (string, []byte, string, error)

/**********/
/* Syncer */
/**********/

type Interface interface {
	Sync(dst string, dstFs afero.Fs, src string, srcFs afero.Fs) error
	SyncProject(prj project.Interface, tmplMgr template.ManagerInterface) error
	SetFileHook(hook FileHookFunc)
	TemplateHook(content interface{}) FileHookFunc
}

func New(logger log.Interface) *syncer {
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

func (snc *syncer) SyncProject(prj project.Interface, tmplMgr template.ManagerInterface) error {
	snc.SetFileHook(snc.TemplateHook(prj.GetOptions()))

	// Get template
	tmpl, err := tmplMgr.Get(prj.GetTemplate())
	if err != nil {
		return err
	}

	for _, unit := range tmpl.GetSync() {
		srcFs := tmpl.GetFs()
		if unit.Template != "" {
			srcTpl, err := tmplMgr.Get(unit.Template)
			if err != nil {
				return err
			}
			srcFs = srcTpl.GetFs()
		}
		err := snc.Sync(unit.Destination, prj.GetFs(), unit.Source, srcFs)
		if err != nil {
			return err
		}
	}

	return nil
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
		}).Debug("Syncing directory...")

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
			if file.IsDir() {
				err = snc.Sync(dstFile, dstFs, srcFile, srcFs)
				if err != nil {
					return err
				}
			} else {
				// Source file info
				srcFileInfo, srcFileErr := srcFs.Stat(srcFile)

				if srcFileErr != nil {
					// Source file does not exist
					if os.IsNotExist(srcFileErr) {
						return &SourceNotExistError{srcFile}
					} else {
						return srcFileErr
					}
				}

				dstFile, err = snc.syncFile(dstFile, dstFs, srcFile, srcFs, srcFileInfo)
				if err != nil {
					return err
				}
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

	_, err := snc.syncFile(dst, dstFs, src, srcFs, srcInfo)

	return err
}

func (snc *syncer) syncFile(dst string, dstFs afero.Fs, src string, srcFs afero.Fs, srcInfo os.FileInfo) (string, error) {
	snc.logger.WithFields(log.Fields{
		"src": src,
		"dst": dst,
	}).Debug("Syncing file...")

	// Content
	srcContent, err := afero.ReadFile(srcFs, src)
	if err != nil {
		return "", err
	}

	// File hook
	if snc.fileHook != nil {
		src, srcContent, dst, err = snc.fileHook(src, srcContent, dst)
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

	srcExecutable := (srcInfo.Mode() & 0100) != 0

	eq, err := snc.equal(dst, dstFs, dstInfo, dstErr, srcContent)
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

		var dstMode os.FileMode = 0666

		// Check user executable permission bit
		if srcExecutable {
			dstMode = 0777
		}

		err := afero.WriteFile(dstFs, dst, srcContent, dstMode)
		if err != nil {
			return "", err
		}

		snc.logger.WithFields(log.Fields{
			"src": src,
			"dst": dst,
		}).Info("File synced")
	}

	// Destination was already existing
	if dstInfo != nil {
		dstMode := dstInfo.Mode()

		dstModeSync := dstMode &^ 0111

		if srcExecutable {
			dstModeSync = dstMode | 0111
		}

		if dstMode != dstModeSync {
			err := dstFs.Chmod(dst, dstModeSync)
			if err != nil {
				return "", err
			}
		}
	}

	return dst, nil
}

func (snc *syncer) equal(dst string, dstFs afero.Fs, dstInfo os.FileInfo, dstErr error, srcContent []byte) (bool, error) {
	// Destination does not exists
	if os.IsNotExist(dstErr) {
		return false, nil
	}

	// Source content and destination file size differs
	if int(dstInfo.Size()) != len(srcContent) {
		return false, nil
	}

	// Checksum differs
	dstContent, err := afero.ReadFile(dstFs, dst)
	if err != nil {
		return false, err
	}

	dstContentHash := md5.Sum(dstContent)
	srcContentHash := md5.Sum(srcContent)

	if dstContentHash != srcContentHash {
		return false, nil
	}

	return true, nil
}

func (snc *syncer) TemplateHook(content interface{}) FileHookFunc {
	return func(src string, srcContent []byte, dst string) (string, []byte, string, error) {
		// Filter on ".tmpl" source files
		if filepath.Ext(src) != ".tmpl" {
			return src, srcContent, dst, nil
		}

		// Remove destination ".tmpl" extension
		dst = strings.TrimSuffix(dst, ".tmpl")

		snc.logger.WithFields(log.Fields{
			"src": src,
			"dst": dst,
		}).Debug("Syncing file template...")

		// Sprig functions
		funcs := sprig.TxtFuncMap()

		// Extra functions
		funcs["toYaml"] = func(v interface{}) string {
			content, err := yaml.Marshal(v)
			if err != nil {
				return ""
			}
			return string(content)
		}

		tmpl, err := engine.New(src).Funcs(funcs).Parse(string(srcContent))
		if err != nil {
			return "", nil, "", err
		}

		var tmplContent bytes.Buffer

		err = tmpl.Execute(&tmplContent, content)
		if err != nil {
			return "", nil, "", err
		}

		srcContent = tmplContent.Bytes()

		return src, srcContent, dst, nil
	}
}
