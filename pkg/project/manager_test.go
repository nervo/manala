package project

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/spf13/afero"
	"reflect"
	"testing"
)

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
		NewFactory(
			logger,
		),
		logger,
	)

	type args struct {
		dir string
	}
	type want struct {
		template string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr error
	}{
		{
			"project",
			args{dir: "project"},
			want{template: "foo"},
			nil,
		},
		{
			"project_parent",
			args{dir: "project_parent/foo"},
			want{template: "foo"},
			nil,
		},
		{
			"project_parent_not_found",
			args{dir: "project_parent_not_found/foo"},
			want{},
			ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prj, err := manager.Find(tt.args.dir)
			if err != tt.wantErr {
				t.Errorf("manager.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != (want{}) {
				got := want{template: prj.GetTemplate()}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("manager.Find() got = %v, want %v", got, tt.want)
				}
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
		NewFactory(
			logger,
		),
		logger,
	)

	type args struct {
		dir string
	}
	type want struct {
		template string
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
			[]want{{template: "foo"}, {template: "bar"}, {template: "baz"}},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make([]want, 0)
			err := manager.Walk(tt.args.dir, func(prj Interface) {
				got = append(got, want{template: prj.GetTemplate()})
			})
			if err != tt.wantErr {
				t.Errorf("manager.Walk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("manager.Walk() got = %v, want %v", got, tt.want)
			}
		})
	}
}
