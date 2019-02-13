package storage

import (
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	for name, c := range map[string]struct {
		path   string
		err    error
		expect Configs
	}{
		"normal": {
			"./testdata/config.json",
			nil,
			Configs{
				{"git", "http://github.com/1/1", "tw", "123"},
				{"local", "/tmp/1/1", "", ""},
				{"local", "tmp/1/1", "", ""},
				{"github", "http://github.com/2/2", "", "321"},
			},
		},

		"nonexisting": {
			"invalid/path/to/config.json",
			pathNotFound,
			nil,
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			cfg, err := getConfig(c.path)
			if e := matchError(c.err, err); e != nil {
				t.Fatal(e)
			}
			if len(cfg) != len(c.expect) {
				t.Errorf("got %d cfg, but want %d\n", len(cfg), len(c.expect))
			}
			for i := 0; i < len(c.expect); i++ {
				if !reflect.DeepEqual(cfg[i], c.expect[i]) {
					t.Errorf("config %d different: got %#v, want %#v\n",
						i, cfg[i], c.expect[i])
				}
			}
		})
	}
}
