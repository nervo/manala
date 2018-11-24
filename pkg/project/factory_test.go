package project

import (
	"github.com/apex/log/handlers/discard"
	"testing"

	"github.com/apex/log"
)

func TestFactory_Create(t *testing.T) {
	// Factory
	factory := &Factory{
		Logger: &log.Logger{
			Handler: discard.Default,
		},
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name         string
		args         args
		wantTemplate string
		wantErr      bool
	}{
		{"project", args{dir: "testdata/factory/project"}, "bar", false},
		{"project_not_found", args{"testdata/factory/project_not_found"}, "", true},
		{"project_template_not_defined", args{"testdata/factory/project_template_not_defined"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := factory.Create(tt.args.dir)
			// Todo: not just error, but error type
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Todo: test for a real Project, and not just its template
			if tt.wantTemplate != "" {
				template := got.GetTemplate()
				if template != tt.wantTemplate {
					t.Errorf("New() = %v, want %v", template, tt.wantTemplate)
				}
			}
		})
	}
}
