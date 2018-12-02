package template

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
		name string
		dir  string
	}
	tests := []struct {
		name    string
		args    args
		want    [3]string
		wantErr error
	}{
		{"template", args{name: "foo", dir: "/template"}, [3]string{"foo", "Foo", "/template"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := factory.Create(tt.args.name, tt.args.dir)
			if err != tt.wantErr {
				t.Errorf("factory.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != [3]string{} {
				got := [3]string{tpl.GetName(), tpl.GetDescription(), tpl.GetDir()}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("factory.Create() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
