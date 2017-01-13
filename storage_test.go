package storage

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"sync"
	"testing"
	"time"
)

func init() {
	debug = false
	Init("src/github.com/tw4452852/storage/testdata/config.xml")
}

type entry struct {
	data string
}

//implement Poster interface
func (e *entry) Key() string {
	return e.data
}
func (e *entry) Date() time.Time {
	return time.Now()
}
func (e *entry) Content() template.HTML {
	return template.HTML("hello test content")
}
func (e *entry) Title() template.HTML {
	return template.HTML("hello test title")
}
func (e *entry) Static(string) io.Reader {
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

type invalidEntry struct {
	data string
}

type testCase struct {
	prepare func() error
	input   []interface{}
	err     error
	checker func(r *Result) error
}

var (
	noKeyerErr = errors.New("arg is not a keyer")
	noFound    = errors.New("can't find want you want")

	ents = []*entry{
		{"1"},
		{"1"},
		{"2"},
	}
	inents = []*invalidEntry{
		{"1"},
		{"1"},
		{"2"},
	}
)

func TestAdd(t *testing.T) {
	cases := []testCase{
		//add
		{
			nil,
			[]interface{}{ents[0]},
			nil,
			func(*Result) error {
				if dataCenter.find(ents[0].data) != ents[0] {
					return errors.New("add valid one failed\n")
				}
				return nil
			},
		},
		{
			nil,
			[]interface{}{ents[1], ents[2]},
			nil,
			func(*Result) error {
				if dataCenter.find(ents[1].data) != ents[1] {
					return errors.New("valid+valid add: first is not found\n")
				}
				if dataCenter.find(ents[2].data) != ents[2] {
					return errors.New("valid+valid add: second is not found\n")
				}
				return nil
			},
		},
	}
	for _, c := range cases {
		dataCenter.reset()
		if c.prepare != nil {
			if err := c.prepare(); err != nil {
				t.Fatal(err)
			}
		}
		inputs := make([]Poster, len(c.input))
		for i, p := range c.input {
			inputs[i] = p.(Poster)
		}
		if e := matchError(c.err, Add(inputs...)); e != nil {
			t.Fatal(e)
		}
		if c.checker != nil {
			if err := c.checker(nil); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestUpdate(t *testing.T) {
	cases := []testCase{
		//update
		{
			nil,
			[]interface{}{ents[0], ents[1]},
			nil,
			func(*Result) error {
				if dataCenter.find(ents[0].data) != ents[1] {
					return errors.New("update valid+valid: not update\n")
				}
				return nil
			},
		},
	}
	for _, c := range cases {
		dataCenter.reset()
		if c.prepare != nil {
			if err := c.prepare(); err != nil {
				t.Fatal(err)
			}
		}
		inputs := make([]Poster, len(c.input))
		for i, p := range c.input {
			inputs[i] = p.(Poster)
		}
		if e := matchError(c.err, Add(inputs...)); e != nil {
			t.Fatal(e)
		}
		if c.checker != nil {
			if err := c.checker(nil); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestRemove(t *testing.T) {
	cases := []testCase{
		//remove
		{
			func() error {
				if err := Add(ents[0], ents[1], ents[2]); err != nil {
					return err
				}
				return nil
			},
			[]interface{}{ents[0]},
			nil,
			func(*Result) error {
				if dataCenter.find(ents[0].data) != nil {
					return errors.New("remove exist one: not remove\n")
				}
				if dataCenter.find(ents[2].data) != ents[2] {
					return errors.New("remove exist one: remove another\n")
				}
				return nil
			},
		},
		{
			nil,
			[]interface{}{ents[0]},
			nil,
			func(*Result) error {
				if dataCenter.find(ents[0].data) != nil {
					return errors.New("remove no exist one: not remove\n")
				}
				return nil
			},
		},
	}
	for _, c := range cases {
		dataCenter.reset()
		if c.prepare != nil {
			if err := c.prepare(); err != nil {
				t.Fatal(err)
			}
		}
		inputs := make([]Keyer, len(c.input))
		for i, v := range c.input {
			inputs[i] = v.(Keyer)
		}
		if e := matchError(c.err, Remove(inputs...)); e != nil {
			t.Fatal(e)
		}
		if c.checker != nil {
			if err := c.checker(nil); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestGet(t *testing.T) {
	cases := []testCase{
		//get
		{
			func() error {
				if err := Add(ents[0], ents[1], ents[2]); err != nil {
					return err
				}
				return nil
			},
			[]interface{}{},
			nil,
			func(r *Result) error {
				if len(r.Content) != 2 {
					return fmt.Errorf("get all: result len(%d) != expect(%d)\n",
						len(r.Content), 2)
				}
				if err := compareTwo(ents[1:], r.Content); err != nil {
					return err
				}
				return nil
			},
		},

		{
			func() error {
				if err := Add(ents[0], ents[1]); err != nil {
					return err
				}
				return nil
			},
			[]interface{}{ents[0]},
			nil,
			func(r *Result) error {
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

		{
			func() error {
				if err := Add(ents[0], ents[1]); err != nil {
					return err
				}
				return nil
			},
			[]interface{}{ents[0], ents[1]},
			nil,
			func(r *Result) error {
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

		{
			func() error {
				if err := Add(ents[0], ents[1]); err != nil {
					return err
				}
				return nil
			},
			[]interface{}{ents[0], ents[2]},
			noFound,
			func(r *Result) error {
				if r != nil {
					return errors.New("add some: result should be nil\n")
				}
				return nil
			},
		},
	}
	for _, c := range cases {
		dataCenter.reset()
		if c.prepare != nil {
			if err := c.prepare(); err != nil {
				t.Fatal(err)
			}
		}
		inputs := make([]Keyer, len(c.input))
		for i, v := range c.input {
			inputs[i] = v.(Keyer)
		}
		result, err := Get(inputs...)
		if e := matchError(c.err, err); e != nil {
			t.Fatal(e)
		}
		if c.checker != nil {
			if err := c.checker(result); err != nil {
				t.Fatal(err)
			}
		}
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
	dataCenter.reset()
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				Add(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}

func BenchmarkAddRemove(b *testing.B) {
	b.StopTimer()
	dataCenter.reset()
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				Add(ents[j])
				Remove(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	dataCenter.reset()
	Add(ents[0])
	Add(ents[1])
	Add(ents[2])
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				Get(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}

func BenchmarkAll(b *testing.B) {
	b.StopTimer()
	dataCenter.reset()
	waiter := &sync.WaitGroup{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		waiter.Add(3)
		for j := 0; j < 3; j++ {
			go func(j int) {
				Add(ents[j])
				Get(ents[j])
				Add(ents[j])
				Remove(ents[j])
				waiter.Done()
			}(j)
		}
		waiter.Wait()
	}
}
