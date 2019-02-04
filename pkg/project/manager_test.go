package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_manager_Create(t *testing.T) {
	// File system
	fs := afero.NewBasePathFs(
		afero.NewOsFs(),
		"testdata/manager",
	)
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Manager
	manager := NewManager(
		fs,
		logger,
	)

	type args struct {
		fs afero.Fs
	}
	type want struct {
		template   string
		repository string
	}
	tests := []struct {
		name    string
		args    args
		want    *want
		wantErr error
	}{
		{
			"project",
			args{fs: afero.NewBasePathFs(fs, "project")},
			&want{template: "foo", repository: ""},
			nil,
		},
		{
			"project_local",
			args{fs: afero.NewBasePathFs(fs, "project_local")},
			&want{template: "bar", repository: ""},
			nil,
		},
		{
			"project_repository",
			args{fs: afero.NewBasePathFs(fs, "project_repository")},
			&want{template: "foo", repository: "foo.git"},
			nil,
		},
		{
			"project_not_found",
			args{fs: afero.NewBasePathFs(fs, "project_not_found")},
			nil,
			ErrNotFound,
		},
		{
			"project_invalid",
			args{fs: afero.NewBasePathFs(fs, "project_invalid")},
			nil,
			ErrConfig,
		},
		{
			"project_template_not_defined",
			args{fs: afero.NewBasePathFs(fs, "project_template_not_defined")},
			nil,
			ErrConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := manager.Create(tt.args.fs)
			assert.IsType(t, tt.wantErr, err)

			if tt.want != nil {
				assert.Equal(t, tt.want.template, prj.GetTemplate())
				assert.Equal(t, tt.want.repository, prj.GetRepository())
			}
		})
	}
}

func Test_manager_Get(t *testing.T) {
	// File system
	fs := afero.NewBasePathFs(
		afero.NewOsFs(),
		"testdata/manager",
	)
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Manager
	manager := NewManager(
		fs,
		logger,
	)

	type args struct {
		dir string
	}
	type want struct {
		template   string
		repository string
	}
	tests := []struct {
		name    string
		args    args
		want    *want
		wantErr error
	}{
		{
			"project",
			args{dir: "project"},
			&want{template: "foo", repository: ""},
			nil,
		},
		{
			"project_not_found",
			args{dir: "project_not_found"},
			nil,
			ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := manager.Get(tt.args.dir)
			assert.IsType(t, tt.wantErr, err)

			if tt.want != nil {
				assert.Equal(t, tt.want.template, prj.GetTemplate())
				assert.Equal(t, tt.want.repository, prj.GetRepository())
			}
		})
	}
}

func Test_manager_Find(t *testing.T) {
	// File system
	fs := afero.NewBasePathFs(
		afero.NewOsFs(),
		"testdata/manager",
	)
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Manager
	manager := NewManager(
		fs,
		logger,
	)

	type args struct {
		dir string
	}
	type want struct {
		template   string
		repository string
	}
	tests := []struct {
		name    string
		args    args
		want    *want
		wantErr error
	}{
		{
			"project",
			args{dir: "project"},
			&want{template: "foo", repository: ""},
			nil,
		},
		{
			"project_parent",
			args{dir: "project_parent/foo"},
			&want{template: "foo", repository: ""},
			nil,
		},
		{
			"project_parent_not_found",
			args{dir: "project_parent_not_found/foo"},
			nil,
			ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := manager.Find(tt.args.dir)
			assert.IsType(t, tt.wantErr, err)

			if tt.want != nil {
				assert.Equal(t, tt.want.template, prj.GetTemplate())
				assert.Equal(t, tt.want.repository, prj.GetRepository())
			}
		})
	}
}

func Test_manager_Walk(t *testing.T) {
	// File system
	fs := afero.NewBasePathFs(
		afero.NewOsFs(),
		"testdata/manager",
	)
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Manager
	manager := NewManager(
		fs,
		logger,
	)

	type args struct {
		dir string
	}
	type want struct {
		template   string
		repository string
	}
	tests := []struct {
		name    string
		args    args
		want    []want
		wantErr error
	}{
		{
			"projects",
			args{dir: "/projects"},
			[]want{
				{template: "foo", repository: ""},
				{template: "bar", repository: ""},
				{template: "baz", repository: ""},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make([]want, 0)
			err := manager.Walk(tt.args.dir, func(prj *ManagedProject) {
				got = append(got, want{template: prj.GetTemplate(), repository: prj.GetRepository()})
			})
			assert.IsType(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
