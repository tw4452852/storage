package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestConfig(t *testing.T) {
	var localPath string
	if runtime.GOOS == "windows" {
		localPath = filepath.Join(os.Getenv("GOPATH"), "/tmp/1/1")
	} else {
		localPath = "/tmp/1/1"
	}
	cases := []struct {
		path   string
		err    error
		expect *Configs
	}{
		{
			"./testdata/config.xml",
			nil,
			&Configs{
				[]*Config{
					{"git", "http://github.com/1/1", "tw", "123"},
					{"local", localPath, "", ""},
					{"local", filepath.Join(os.Getenv("GOPATH"), "/tmp/1/1"), "", ""},
					{"github", "http://github.com/2/2", "", "321"},
				},
			},
		},

		{
			"invalid/path/to/config.xml",
			pathNotFound,
			nil,
		},
	}
	for _, c := range cases {
		cfg, err := getConfig(c.path)
		if e := matchError(c.err, err); e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(cfg, c.expect) {
			t.Errorf("expect configs: %v\n, real configs %v\n",
				*c.expect, *cfg)
		}
	}
}
