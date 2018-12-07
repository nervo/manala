package repository

import (
	"errors"
	"github.com/spf13/afero"
)

var (
	ErrUnclonable = errors.New("repository unclonable")
	ErrUnopenable = errors.New("repository unopenable")
	ErrInvalid    = errors.New("repository invalid")
)

type Interface interface {
	GetSrc() string
	GetFs() afero.Fs
}

type repository struct {
	src string
	fs  afero.Fs
}

func (rep *repository) GetSrc() string {
	return rep.src
}

func (rep *repository) GetFs() afero.Fs {
	return rep.fs
}
