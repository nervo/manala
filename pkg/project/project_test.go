package project

import (
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name         string
		args         args
		wantTemplate string
		wantErr      bool
	}{
		{"ok", args{"testdata/ok"}, "bar", false},
		{"not_found", args{"testdata/not_found"}, "", true},
		{"template_not_defined", args{"testdata/template_not_defined"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.dir)
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
