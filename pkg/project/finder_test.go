package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"testing"
)

func TestFinder_Find(t *testing.T) {
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Finder
	finder := NewFinder(
		NewFactory(
			afero.NewBasePathFs(afero.NewOsFs(), "testdata/finder"),
			logger,
		),
		logger,
	)
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
		{"project_parent", args{dir: "/project_parent/foo"}, "foo", nil},
		{"project_parent_not_found", args{dir: "/project_parent_not_found/foo"}, "", ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := finder.Find(tt.args.dir)
			if err != tt.wantErr {
				t.Errorf("finder.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Todo: test for a real project, and not just its template
			if tt.wantTemplate != "" {
				template := got.GetTemplate()
				if template != tt.wantTemplate {
					t.Errorf("finder.Find() template = %v, wantTemplate %v", template, tt.wantTemplate)
				}
			}
		})
	}
}
