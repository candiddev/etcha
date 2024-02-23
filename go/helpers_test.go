package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func TestDirFileRunParse(t *testing.T) {
	root := uint32(0)
	permissions := fs.FileMode(uint32(256))
	permissionsDir := fs.FileMode(uint32(256) + uint32(fs.ModeDir))

	tests := map[string]struct {
		dir             bool
		group           []string
		owner           []string
		permissions     []string
		wantErr         bool
		wantGroup       *uint32
		wantOwner       *uint32
		wantPermissions *fs.FileMode
	}{
		"all values": {
			group:           []string{"root"},
			owner:           []string{"root"},
			permissions:     []string{"0400"},
			wantGroup:       &root,
			wantOwner:       &root,
			wantPermissions: &permissions,
		},
		"all values dir": {
			dir:             true,
			group:           []string{"root"},
			owner:           []string{"root"},
			permissions:     []string{"0400"},
			wantGroup:       &root,
			wantOwner:       &root,
			wantPermissions: &permissionsDir,
		},
		"no values": {},
		"wrong group": {
			group:   []string{"not real"},
			wantErr: true,
		},
		"wrong owner": {
			owner:   []string{"not real"},
			wantErr: true,
		},
		"wrong permissions": {
			permissions: []string{"wrong"},
			wantErr:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, o, g, err := dirFileRunParse(tc.dir, cli.Flags{
				"g": {
					Values: tc.group,
				},
				"o": {
					Values: tc.owner,
				},
				"p": {
					Values: tc.permissions,
				},
			})
			assert.Equal(t, err != nil, tc.wantErr)
			assert.Equal(t, g, tc.wantGroup)
			assert.Equal(t, o, tc.wantOwner)
			assert.Equal(t, p, tc.wantPermissions)
		})
	}
}

func TestDirFileRunMk(t *testing.T) {
	p600 := fs.FileMode(0600)
	p700 := fs.FileMode(0700)

	tests := []struct {
		change          bool
		dir             bool
		name            string
		path            string
		permissions     *fs.FileMode
		wantErr         bool
		wantPermissions fs.FileMode
	}{
		{
			name:    "missing_file_check1",
			path:    "test1",
			wantErr: true,
		},
		{
			name:            "missing_file_change_noperm",
			change:          true,
			path:            "test1",
			wantPermissions: fs.FileMode(0644),
		},
		{
			name:            "missing_file_check2",
			path:            "test1",
			wantPermissions: fs.FileMode(0644),
		},
		{
			name:            "missing_file_change_perm",
			change:          true,
			path:            "test2",
			permissions:     &p600,
			wantPermissions: p600,
		},
		{
			name:    "missing_dir_check1",
			dir:     true,
			path:    "testdir1",
			wantErr: true,
		},
		{
			name:            "missing_dir_change_noperm",
			change:          true,
			dir:             true,
			path:            "testdir1",
			wantPermissions: fs.ModeDir + 0755,
		},
		{
			name:            "missing_dir_check2",
			dir:             true,
			path:            "testdir1",
			wantPermissions: fs.ModeDir + 0755,
		},
		{
			name:            "missing_dir_change_perm",
			change:          true,
			dir:             true,
			path:            "testdir2",
			permissions:     &p700,
			wantPermissions: fs.ModeDir + 0700,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := dirFileRunMk(tc.change, tc.dir, []byte("hello"), tc.path, tc.permissions)

			assert.Equal(t, err != nil, tc.wantErr)
			if !tc.wantErr {
				assert.Equal(t, f != nil, true)
				assert.Equal(t, f.Mode(), tc.wantPermissions)

				if !tc.dir {
					o, _ := os.ReadFile(tc.path)
					assert.Equal(t, string(o), "hello")
				}
			}
		})
	}

	os.RemoveAll("test1")
	os.RemoveAll("test2")
	os.RemoveAll("testdir1")
	os.RemoveAll("testdir2")
}

func TestDirFileRun(t *testing.T) {
	c := config.Default()

	ctx := context.Background()
	ctx = logger.SetNoColor(ctx, true)

	gid := strconv.Itoa(os.Getgid())
	uid := strconv.Itoa(os.Getuid())

	tests := []struct {
		args        []string
		dir         bool
		group       []string
		name        string
		owner       []string
		permissions []string
		stdin       string
		want        string
	}{
		{
			name: "path_check1",
			args: []string{
				"check",
				"test",
			},
			want: "ERROR stat test: no such file or directory\n",
		},
		{
			name: "path_change",
			args: []string{
				"change",
				"test",
			},
		},
		{
			name: "path_check2",
			args: []string{
				"change",
				"test",
			},
		},
		{
			name: "path_remove1",
			args: []string{
				"remove",
				"test",
			},
		},
		{
			name: "path_remove2",
			args: []string{
				"remove",
				"test",
			},
		},
		{
			name: "path_check3",
			args: []string{
				"check",
				"test",
			},
			want: "ERROR stat test: no such file or directory\n",
		},
		{
			name: "path_bad",
			args: []string{
				"change",
				"/root/test",
			},
			want: "ERROR stat /root/test: permission denied\n",
		},
		{
			name: "full_change_name",
			args: []string{
				"change",
				"test",
			},
			group:       []string{gid},
			owner:       []string{uid},
			permissions: []string{"600"},
		},
		{
			name: "full_change_id",
			args: []string{
				"change",
				"test",
			},
			group:       []string{gid},
			owner:       []string{uid},
			permissions: []string{"600"},
		},
		{
			name: "full_check_id1",
			args: []string{
				"check",
				"test",
			},
			group:       []string{gid},
			owner:       []string{uid},
			permissions: []string{"600"},
		},
		{
			name: "full_check_id2",
			args: []string{
				"check",
				"test",
			},
			group:       []string{gid},
			owner:       []string{uid},
			permissions: []string{"640"},
			want:        "ERROR file test does not match:\n\tmismatch permissions: got 600, want 640\n",
		},
		{
			name: "full_check_id3",
			args: []string{
				"check",
				"test",
			},
			group:       []string{"0"},
			owner:       []string{"0"},
			permissions: []string{"640"},
			want: fmt.Sprintf(`ERROR file test does not match:
	mismatch group: got %s, want 0
	mismatch owner: got %s, want 0
	mismatch permissions: got 600, want 640
`, gid, uid),
		},
		{
			name: "full_check_contents_1",
			args: []string{
				"check",
				"test",
				"-",
			},
			stdin: "contents",
			want: `ERROR file test does not match:
	mismatch contents
`,
		},
		{
			name: "full_change_contents",
			args: []string{
				"change",
				"test",
				"-",
			},
			stdin: "contents",
		},
		{
			name: "full_check_contents_2",
			args: []string{
				"check",
				"test",
				"-",
			},
			stdin: "contents",
		},
		{
			name: "full_change_group",
			args: []string{
				"change",
				"test",
			},
			group: []string{gid},
		},
		{
			name: "full_check_group",
			args: []string{
				"check",
				"test",
			},
			group: []string{gid},
		},
		{
			name: "full_change_permissions",
			args: []string{
				"change",
				"test",
			},
			permissions: []string{"600"},
		},
		{
			name: "full_change_contents",
			args: []string{
				"change",
				"test",
				"hello",
			},
		},
		{
			name: "full_check_contents",
			args: []string{
				"check",
				"test",
				"hello",
			},
		},
		{
			name: "full_remove_id",
			args: []string{
				"remove",
				"test",
			},
			group:       []string{gid},
			owner:       []string{uid},
			permissions: []string{"600"},
		},
		{
			name: "full_check_id4",
			args: []string{
				"check",
				"test",
			},
			want:        "ERROR stat test: no such file or directory\n",
			group:       []string{gid},
			owner:       []string{uid},
			permissions: []string{"600"},
		},
		{
			dir:  true,
			name: "check dir1",
			args: []string{
				"check",
				"testdata",
			},
			want:        "ERROR stat testdata: no such file or directory\n",
			permissions: []string{"0700"},
		},
		{
			dir:  true,
			name: "change dir1",
			args: []string{
				"change",
				"testdata",
			},
			permissions: []string{"0700"},
		},
		{
			dir:  true,
			name: "check dir2",
			args: []string{
				"check",
				"testdata",
			},
			permissions: []string{"0700"},
		},
		{
			dir:  true,
			name: "change dir2",
			args: []string{
				"change",
				"testdata",
			},
			permissions: []string{"0700"},
		},
		{
			dir:  true,
			name: "remove dir",
			args: []string{
				"remove",
				"testdata",
			},
		},
		{
			dir:  true,
			name: "check dir1",
			args: []string{
				"check",
				"testdata",
			},
			permissions: []string{"0700"},
			want:        "ERROR stat testdata: no such file or directory\n",
		},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			logger.SetStd()
			cli.SetStdin(tests[i].stdin)

			cmd := "file"
			if tests[i].dir {
				cmd = "dir"
			}

			f := cli.Flags{
				"g": {
					Values: tests[i].group,
				},
				"o": {
					Values: tests[i].owner,
				},
				"p": {
					Values: tests[i].permissions,
				},
			}

			if tests[i].want == "" {
				assert.HasErr(t, file.Run(ctx, append([]string{cmd}, tests[i].args...), f, c), nil)
			} else {
				assert.HasErr(t, file.Run(ctx, append([]string{cmd}, tests[i].args...), f, c), errs.ErrReceiver)
			}

			assert.Equal(t, logger.ReadStd(), tests[i].want)
		})
	}
}
