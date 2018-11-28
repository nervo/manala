package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"reflect"
	"testing"
)

func Test_finder_Find(t *testing.T) {
	// File system
	fs := afero.NewBasePathFs(afero.NewOsFs(), "testdata/finder")
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Finder
	finder := NewFinder(
		fs,
		NewFactory(fs, logger),
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
		{"project_parent", args{dir: "/project_parent/foo"}, [2]string{"/project_parent", "foo"}, nil},
		{"project_parent_not_found", args{dir: "/project_parent_not_found/foo"}, [2]string{}, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := finder.Find(tt.args.dir)
			if err != tt.wantErr {
				t.Errorf("finder.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != [2]string{} {
				got := [2]string{prj.GetDir(), prj.GetTemplate()}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("finder.Find() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_finder_Walk(t *testing.T) {
	// File system
	fs := afero.NewBasePathFs(afero.NewOsFs(), "testdata/finder")
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Finder
	finder := NewFinder(
		fs,
		NewFactory(fs, logger),
		logger,
	)
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    [][2]string
		wantErr error
	}{
		{"projects", args{dir: "/projects"}, [][2]string{{"/projects", "foo"}, {"/projects/bar", "bar"}, {"/projects/baz/baz", "baz"}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make([][2]string, 0)
			err := finder.Walk(tt.args.dir, func(prj Interface) {
				got = append(got, [2]string{prj.GetDir(), prj.GetTemplate()})
			})
			if err != tt.wantErr {
				t.Errorf("finder.Walk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("finder.Walk() got = %v, want %v", got, tt.want)
			}
		})
	}
}
