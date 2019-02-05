package syncer

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
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
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Syncer
	snc := &syncer{
		delete: true,
		logger: logger,
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
		want    *want
		wantErr error
	}{
		{
			"source_not_exist",
			args{dst: "baz", src: "baz"},
			nil,
			&SourceNotExistError{},
		},
		{
			"file_not_exist",
			args{dst: "foo", src: "foo"},
			&want{file: "foo", content: "bar"},
			nil,
		},
		{
			"file_exist_same",
			args{dst: "file_bar", src: "foo"},
			&want{file: "file_bar", content: "bar"},
			nil,
		},
		{
			"file_exist_differs",
			args{dst: "file_foo", src: "foo"},
			&want{file: "file_foo", content: "bar"},
			nil,
		},
		{
			"source_file_over_destination_directory_empty",
			args{dst: "dir_empty", src: "foo"},
			&want{file: "dir_empty", content: "bar"},
			nil,
		},
		{
			"source_file_over_destination_directory",
			args{dst: "dir", src: "foo"},
			&want{file: "dir", content: "bar"},
			nil,
		},
		{
			"directory_not_exist",
			args{dst: "bar", src: "bar"},
			&want{file: "bar/foo", content: "baz"},
			nil,
		},
		{
			"directory_exist",
			args{dst: "dir", src: "bar"},
			&want{file: "dir/foo", content: "baz"},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Destination file system
			dstFs := afero.NewMemMapFs()
			_ = afero.WriteFile(dstFs, "file_foo", []byte("foo"), 0666)
			_ = afero.WriteFile(dstFs, "file_bar", []byte("bar"), 0666)
			_ = dstFs.Mkdir("dir_empty", 0755)
			_ = dstFs.Mkdir("dir", 0755)
			_, _ = dstFs.Create("dir/foo")
			_ = afero.WriteFile(dstFs, "dir/foo", []byte("bar"), 0666)
			_ = dstFs.Mkdir("dir/bar", 0755)
			_, _ = dstFs.Create("dir/bar/foo")

			err := snc.Sync(tt.args.dst, dstFs, tt.args.src, srcFs)
			assert.IsType(t, tt.wantErr, err)

			if tt.want != nil {
				exists, _ := afero.Exists(dstFs, tt.want.file)
				assert.True(t, exists)
				content, _ := afero.ReadFile(dstFs, tt.want.file)
				assert.Equal(t, tt.want.content, string(content))
			}
		})
	}
}

func Test_syncer_Sync_executable(t *testing.T) {
	// Source file system
	srcFs := afero.NewBasePathFs(
		afero.NewOsFs(),
		"testdata/fs",
	)
	// Logger
	logger := &log.Logger{
		Handler: discard.Default,
	}
	// Syncer
	snc := &syncer{
		delete: true,
		logger: logger,
	}

	type args struct {
		dst string
		src string
	}
	tests := []struct {
		name           string
		args           args
		wantExecutable bool
	}{
		{
			"executable_true",
			args{dst: "executable", src: "executable_true"},
			true,
		},
		{
			"executable_false",
			args{dst: "executable", src: "executable_false"},
			false,
		},
		{
			"executable_false_false",
			args{dst: "executable_false", src: "executable_false"},
			false,
		},
		{
			"executable_false_true",
			args{dst: "executable_false", src: "executable_true"},
			true,
		},
		{
			"executable_true_false",
			args{dst: "executable_true", src: "executable_false"},
			false,
		},
		{
			"executable_false_true",
			args{dst: "executable_true", src: "executable_true"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Destination file system
			dstFs := afero.NewMemMapFs()
			_ = afero.WriteFile(dstFs, "executable_false", []byte(""), 0666)
			_ = afero.WriteFile(dstFs, "executable_true", []byte(""), 0777)

			err := snc.Sync(tt.args.dst, dstFs, tt.args.src, srcFs)
			assert.IsType(t, nil, err)

			dstInfo, _ := dstFs.Stat(tt.args.dst)
			assert.Equal(t, tt.wantExecutable, (dstInfo.Mode() & 0100) != 0)
		})
	}
}
