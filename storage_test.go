package storage

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"testing"
	"time"
)

func init() {
	debug = false
}

type entry struct {
	data string
}

// implement Poster interface
func (e *entry) Key() string {
	return e.data
}
func (e *entry) Date() time.Time {
	return parseTime("2018-10-" + e.data)
}
func (e *entry) Content() string {
	return "hello test content"
}
func (e *entry) Title() string {
	return "hello test title"
}
func (e *entry) Static(string) io.ReadCloser {
	return nil
}
func (e *entry) Update() error {
	return nil
}
func (e *entry) Tags() []string {
	return nil
}
func (e *entry) IsSlide() bool {
	return false
}
func (e *entry) StaticList() []string {
	return nil
}

type testCase struct {
	prepare func() error
	input   []interface{}
	err     error
	checker func(r *Result) error
}

var (
	ents = []*entry{
		{"01"},
		{"01"},
		{"02"},
	}
)

func TestStorageAdd(t *testing.T) {
	for name, c := range map[string]struct {
		input   []Poster
		checker func(s *Storage) error
	}{
		"addOne": {
			input: []Poster{ents[0]},
			checker: func(s *Storage) error {
				if s.data[ents[0].data] != ents[0] {
					return errors.New("add valid one failed\n")
				}
				return nil
			},
		},
		"addTwo": {
			input: []Poster{ents[1], ents[2]},
			checker: func(s *Storage) error {
				if s.data[ents[1].data] != ents[1] {
					return errors.New("valid+valid add: first is not found\n")
				}
				if s.data[ents[2].data] != ents[2] {
					return errors.New("valid+valid add: second is not found\n")
				}
				return nil
			},
		},
		"updateOne": {
			input: []Poster{ents[0], ents[1]},
			checker: func(s *Storage) error {
				if s.data[ents[0].data] != ents[1] {
					return errors.New("update valid+valid: not update\n")
				}
				return nil
			},
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			s, err := New("./testdata/repos.json")
			if err != nil {
				t.Fatal(err)
			}
			err = s.Add(c.input...)
			if err != nil {
				t.Fatal(err)
			}
			err = c.checker(s)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestStorageRemove(t *testing.T) {
	for name, c := range map[string]struct {
		prepare func(s *Storage) error
		input   []Keyer
		checker func(s *Storage) error
	}{
		"removeOne": {
			prepare: func(s *Storage) error {
				if err := s.Add(ents[0], ents[1], ents[2]); err != nil {
					return err
				}
				return nil
			},
			input: []Keyer{ents[0]},
			checker: func(s *Storage) error {
				if s.data[ents[0].data] != nil {
					return errors.New("remove exist one: not remove\n")
				}
				if s.data[ents[2].data] != ents[2] {
					return errors.New("remove exist one: remove another\n")
				}
				return nil
			},
		},
		"removeNotExist": {
			input: []Keyer{ents[0]},
			checker: func(s *Storage) error {
				if s.data[ents[0].data] != nil {
					return errors.New("remove no exist one: not remove\n")
				}
				return nil
			},
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			s, err := New("./testdata/repos.json")
			if err != nil {
				t.Fatal(err)
			}
			if c.prepare != nil {
				if err = c.prepare(s); err != nil {
					t.Fatal(err)
				}
			}
			if err = s.Remove(c.input...); err != nil {
				t.Fatal(err)
			}
			if err = c.checker(s); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestStorageGet(t *testing.T) {
	for name, c := range map[string]struct {
		prepare   func(*Storage) error
		input     []Keyer
		expectErr error
		checker   func(*Result) error
	}{
		"getAll": {
			prepare: func(s *Storage) error {
				if err := s.Add(ents[0], ents[1], ents[2]); err != nil {
					return err
				}
				return nil
			},
			input: []Keyer{},
			checker: func(r *Result) error {
				if len(r.Content) != 2 {
					return fmt.Errorf("get all: result len(%d) != expect(%d)\n",
						len(r.Content), 2)
				}
				sort.Sort(r)
				if r.Content[0] != ents[2] ||
					r.Content[1] != ents[1] {
					return fmt.Errorf("not expected result: %#v\n", r)
				}
				return nil
			},
		},

		"getOne": {
			prepare: func(s *Storage) error {
				if err := s.Add(ents[0], ents[1]); err != nil {
					return err
				}
				return nil
			},
			input: []Keyer{ents[0]},
			checker: func(r *Result) error {
				if len(r.Content) != 1 {
					return fmt.Errorf("get some: result len(%d) != expect(%d)\n",
						len(r.Content), 1)
				}
				if r.Content[0] != ents[1] {
					return noFound
				}
				return nil
			},
		},

		"getTwo": {
			prepare: func(s *Storage) error {
				if err := s.Add(ents[0], ents[1]); err != nil {
					return err
				}
				return nil
			},
			input: []Keyer{ents[0], ents[1]},
			checker: func(r *Result) error {
				if len(r.Content) != 2 {
					return fmt.Errorf("get some: result len(%d) != expect(%d)\n",
						len(r.Content), 2)
				}
				if r.Content[0] != ents[1] || r.Content[1] != ents[1] {
					return noFound
				}
				return nil
			},
		},

		"getNotExist": {
			prepare: func(s *Storage) error {
				if err := s.Add(ents[0], ents[1]); err != nil {
					return err
				}
				return nil
			},
			input:     []Keyer{ents[0], ents[2]},
			expectErr: noFound,
			checker: func(r *Result) error {
				if r != nil {
					return errors.New("add some: result should be nil\n")
				}
				return nil
			},
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			s, err := New("./testdata/repos.json")
			if err != nil {
				t.Fatal(err)
			}
			if c.prepare != nil {
				if err = c.prepare(s); err != nil {
					t.Fatal(err)
				}
			}
			r, err := s.Get(c.input...)
			if err != c.expectErr {
				t.Fatalf("expect error: %v, but got %v\n", c.expectErr, err)
			}
			if err = c.checker(r); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func compareTwo(expects []*entry, reals []Poster) error {
check:
	for _, expect := range expects {
		for _, real := range reals {
			if expect == real {
				continue check
			}
		}
		return fmt.Errorf("get all: expect %v not in result\n",
			expect)
	}
	return nil
}

func BenchmarkAddUpdate(b *testing.B) {
	b.StopTimer()
	s, err := New("./testdata/repos.json")
	if err != nil {
		b.Fatal(err)
	}
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				s.Add(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}

func BenchmarkAddRemove(b *testing.B) {
	b.StopTimer()
	s, err := New("./testdata/repos.json")
	if err != nil {
		b.Fatal(err)
	}
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				s.Add(ents[j])
				s.Remove(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	s, err := New("./testdata/repos.json")
	if err != nil {
		b.Fatal(err)
	}
	s.Add(ents[0])
	s.Add(ents[1])
	s.Add(ents[2])
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				s.Get(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}

func BenchmarkAll(b *testing.B) {
	b.StopTimer()
	s, err := New("./testdata/repos.json")
	if err != nil {
		b.Fatal(err)
	}
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				s.Add(ents[j])
				s.Get(ents[j])
				s.Add(ents[j])
				s.Remove(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}
