package sync

import (
	"errors"
	"github.com/mostafah/fsync"
	"manala/pkg/project"
	"manala/pkg/template"
	"path/filepath"
)

var (
	Err = errors.New("sync failed")
)

type Interface interface {
	Sync(prj project.Interface, tpl template.Interface) error
}

func NewSync() Interface {
	syncer := fsync.NewSyncer()
	syncer.Delete = true

	return &sync{
		syncer: syncer,
	}
}

type sync struct {
	syncer *fsync.Syncer
}

func (s *sync) Sync(prj project.Interface, tpl template.Interface) error {
	err := s.syncer.Sync(filepath.Join(prj.GetDir(), ".manala"), filepath.Join(tpl.GetDir(), ".manala"))
	if err != nil {
		return Err
	}

	return nil
}
