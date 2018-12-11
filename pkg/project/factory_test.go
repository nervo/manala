package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"reflect"
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
		template string
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
			"project_template_not_defined",
			args{fs: afero.NewBasePathFs(fs, "project_template_not_defined")},
			want{},
			ErrConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := factory.Create(tt.args.fs)
			if err != tt.wantErr {
				t.Errorf("factory.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.want, want{}) {
				got := want{template: prj.GetTemplate(), repository: prj.GetRepository()}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("factory.Create() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
