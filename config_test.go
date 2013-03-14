package storage

import (
	"os"
	"path/filepath"
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
				[]Config{
					{"git", "http://github.com/1/1"},
					{"local", localPath},
					{"local", filepath.Join(os.Getenv("GOPATH"), "/tmp/1/1")},
					{"github", "http://github.com/2/2"},
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
		if checkConfigs(c.expect, cfg, t) {
			t.Errorf("expect configs: %v\n, real configs %v\n",
				*c.expect, *cfg)
		}
	}
}

func checkConfigs(expect, real *Configs, t *testing.T) (needShow bool) {
	if expect != real {
		if expect == nil {
			t.Errorf("expect configs(<nil>),  but get %v\n", *real)
			return false
		}
		if real == nil {
			t.Errorf("expect configs(%v), but get <nil>\n", *expect)
			return false
		}
		if len(expect.Content) != len(real.Content) {
			t.Errorf("expect %d content, but get %d\n",
				len(expect.Content), len(real.Content))
			return true
		}
		for i, ec := range expect.Content {
			if real.Content[i] != ec {
				t.Errorf("expect %dth content(%v), but get(%v)\n",
					i, ec, real.Content[i])
				return true
			}
		}
	}
	return false
}
