package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

type nopStorage struct{}

var (
	_            Storager = nopStorage{}
	pathNotFound          = errors.New("no such file or directory")
)

func (nopStorage) Add(...Poster) error           { return nil }
func (nopStorage) Destroy()                      {}
func (nopStorage) Get(...Keyer) (*Result, error) { return nil, nil }
func (nopStorage) Remove(...Keyer) error         { return nil }

func matchError(expect, real error) error {
	if expect != real {
		if expect == nil {
			return fmt.Errorf("expect err(nil), but get err(%s)\n", real.Error())
		}
		if real == nil {
			return fmt.Errorf("expect err(%s), but get err(nil)\n", expect.Error())
		}
		if strings.Contains(real.Error(), expect.Error()) {
			return nil
		}
		return fmt.Errorf("expect err(%s), but get err(%s)\n",
			expect.Error(), real.Error())
	}
	return nil
}

func isPosterEqual(a, b Poster) bool {
	if a == b {
		return true
	}

	if a.Date() != b.Date() {
		return false
	}
	if a.Content() != b.Content() {
		return false
	}
	if a.Title() != b.Title() {
		return false
	}
	if a.Key() != b.Key() {
		return false
	}
	if !reflect.DeepEqual(a.Tags(), b.Tags()) {
		return false
	}
	if a.IsSlide() != b.IsSlide() {
		return false
	}
	if !reflect.DeepEqual(a.StaticList(), b.StaticList()) {
		return false
	}

	return true
}
func parseTime(v string) time.Time {
	t, err := time.Parse(timePattern, v)
	if err != nil {
		fmt.Println(err)
	}
	return t
}

type testStaticer struct{}

var (
	ts = testStaticer{}
)

func (ts testStaticer) Static(path string) io.ReadCloser {
	path = filepath.Join("./testdata/localRepo", path)
	rc, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	return rc
}
