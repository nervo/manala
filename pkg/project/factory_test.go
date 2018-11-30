package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"reflect"
	"testing"
)

func Test_factory_Create(t *testing.T) {
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Factory
	factory := NewFactory(
		afero.NewBasePathFs(afero.NewOsFs(), "testdata/factory"),
		logger,
	)
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    [2]string
		wantErr error
	}{
		{"project", args{dir: "/project"}, [2]string{"/project", "foo"}, nil},
		{"project_not_found", args{"/project_not_found"}, [2]string{}, ErrNotFound},
		{"project_template_not_defined", args{"/project_template_not_defined"}, [2]string{}, ErrTemplateNotDefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := factory.Create(tt.args.dir)
			if err != tt.wantErr {
				t.Errorf("factory.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != [2]string{} {
				got := [2]string{prj.GetDir(), prj.GetTemplate()}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("factory.Create() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
