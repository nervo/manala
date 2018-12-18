package template

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
		name string
		fs   afero.Fs
	}
	type want struct {
		name        string
		description string
		sync        []syncUnit
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr error
	}{
		{
			"template",
			args{name: "foo", fs: afero.NewBasePathFs(fs, "template")},
			want{name: "foo", description: "Foo", sync: nil},
			nil,
		},
		{
			"template_local",
			args{name: "foo", fs: afero.NewBasePathFs(fs, "template_local")},
			want{name: "foo", description: "Bar", sync: nil},
			nil,
		},
		{
			"template_sync",
			args{name: "foo", fs: afero.NewBasePathFs(fs, "template_sync")},
			want{name: "foo", description: "Foo", sync: []syncUnit{
				{Source: "foo", Destination: "foo"},
				{Source: "foo", Destination: "bar"},
				{Source: "bar", Destination: "bar", Template: "foo"},
				{Source: "bar", Destination: "baz", Template: "foo"},
			}},
			nil,
		},
		{
			"template_not_found",
			args{name: "foo", fs: afero.NewBasePathFs(fs, "template_not_found")},
			want{},
			ErrNotFound,
		},
		{
			"template_invalid",
			args{name: "foo", fs: afero.NewBasePathFs(fs, "template_invalid")},
			want{},
			ErrConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := factory.Create(tt.args.name, tt.args.fs)
			assert.IsType(t, tt.wantErr, err)

			if err == nil {
				assert.Equal(t, tt.want.name, tpl.GetName())
				assert.Equal(t, tt.want.description, tpl.GetDescription())
				assert.Equal(t, tt.want.sync, tpl.GetSync())
			}
		})
	}
}
