package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_factory_Create(t *testing.T) {
	// File system
	fs := afero.NewBasePathFs(
		afero.NewOsFs(),
		"testdata/factory",
	)
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Factory
	factory := NewFactory(
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
		want    want
		wantErr error
	}{
		{
			"project",
			args{fs: afero.NewBasePathFs(fs, "project")},
			want{template: "foo", repository: ""},
			nil,
		},
		{
			"project_local",
			args{fs: afero.NewBasePathFs(fs, "project_local")},
			want{template: "bar", repository: ""},
			nil,
		},
		{
			"project_repository",
			args{fs: afero.NewBasePathFs(fs, "project_repository")},
			want{template: "foo", repository: "foo.git"},
			nil,
		},
		{
			"project_not_found",
			args{fs: afero.NewBasePathFs(fs, "project_not_found")},
			want{},
			ErrNotFound,
		},
		{
			"project_invalid",
			args{fs: afero.NewBasePathFs(fs, "project_invalid")},
			want{},
			ErrConfig,
		},
		{
			"project_template_not_defined",
			args{fs: afero.NewBasePathFs(fs, "project_template_not_defined")},
			want{},
			ErrConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := factory.Create(tt.args.fs)
			assert.IsType(t, tt.wantErr, err)

			if err == nil {
				assert.Equal(t, tt.want.template, prj.GetTemplate())
				assert.Equal(t, tt.want.repository, prj.GetRepository())
			}
		})
	}
}
