package template

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
		name string
		fs   afero.Fs
	}
	type want struct {
		name        string
		description string
		sync        []string
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
			want{name: "foo", description: "Foo", sync: []string{"foo", "bar"}},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := factory.Create(tt.args.name, tt.args.fs)
			if err != tt.wantErr {
				t.Errorf("factory.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.want, want{}) {
				got := want{name: tpl.GetName(), description: tpl.GetDescription(), sync: tpl.GetSync()}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("factory.Create() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
