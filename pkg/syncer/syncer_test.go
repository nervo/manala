package syncer

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_syncer_Sync(t *testing.T) {
	// Source file system
	srcFs := afero.NewBasePathFs(
		afero.NewOsFs(),
		"testdata/fs",
	)
	// Destination file system
	dstFs := afero.NewMemMapFs()
	_ = dstFs.Mkdir("dir", 0755)
	_, _ = dstFs.Create("dir/foo")
	_, _ = dstFs.Create("dir/bar")
	// Syncer
	snc := &syncer{
		delete:  true,
		noTimes: false,
	}

	type args struct {
		dst string
		src string
	}
	type want struct {
		file    string
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr error
	}{
		{
			"source_not_exist",
			args{dst: "bar", src: "bar"},
			want{},
			&SourceNotExistError{},
		},
		{
			"source_file_over_destination_directory",
			args{dst: "dir", src: "foo"},
			want{file: "foo", content: "bar"},
			&SourceFileOverDestinationDirectoryError{},
		},
		{
			"file",
			args{dst: "foo", src: "foo"},
			want{file: "foo", content: "bar"},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := snc.Sync(tt.args.dst, dstFs, tt.args.src, srcFs)
			assert.IsType(t, tt.wantErr, err)

			if err == nil {
				exists, _ := afero.Exists(dstFs, tt.want.file)
				assert.True(t, exists)
				content, _ := afero.ReadFile(dstFs, tt.want.file)
				assert.Equal(t, tt.want.content, string(content))
			}
		})
	}
}
