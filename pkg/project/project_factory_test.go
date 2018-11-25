package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"testing"
)

func TestFactory_Create(t *testing.T) {
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Factory
	factory := &Factory{
		Fs:     afero.NewBasePathFs(afero.NewOsFs(), "testdata/project_factory"),
		Logger: logger,
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name         string
		args         args
		wantTemplate string
		wantErr      error
	}{
		{"project", args{dir: "/project"}, "foo", nil},
		{"project_not_found", args{"/project_not_found"}, "", ErrNotFound},
		{"project_template_not_defined", args{"/project_template_not_defined"}, "", ErrTemplateNotDefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := factory.Create(tt.args.dir)
			if err != tt.wantErr {
				t.Errorf("Factory.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Todo: test for a real Project, and not just its template
			if tt.wantTemplate != "" {
				template := got.GetTemplate()
				if template != tt.wantTemplate {
					t.Errorf("Factory.Create() template = %v, wantTemplate %v", template, tt.wantTemplate)
				}
			}
		})
	}
}
