package storage

import (
	"errors"
)

type Storager interface {
	// Add post into storage, replace if any.
	Add(args ...Poster) error
	// Get post according to the passed key.
	Get(args ...Keyer) (*Result, error)
	// Remove post according to the passed key.
	Remove(args ...Keyer) error
	// Destroy this storage
	Destroy()
}

var _ Storager = &Storage{}

type Storage struct {
	requestCh chan *request     // for outcoming request
	closeCh   chan struct{}     // for exit
	data      map[string]Poster // internal data storage
}

func New(configPath string) (*Storage, error) {
	s := &Storage{
		requestCh: make(chan *request),
		closeCh:   make(chan struct{}),
		data:      make(map[string]Poster),
	}
	go s.serve()

	_, err := newRepos(configPath, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (d *Storage) serve() {
	for {
		select {
		case req := <-d.requestCh:
			d.handleRequest(req)
		case <-d.closeCh:
			return
		}
	}
}

var noFound = errors.New("can't find what you want")

func (d *Storage) handleRequest(req *request) {
	loopArgs := func(action func(key string, arg interface{}) error) error {
		for _, arg := range req.args {
			// Only accept the things implement keyer
			if keyer, ok := arg.(Keyer); ok {
				key := keyer.Key()
				if err := action(key, arg); err != nil {
					return err
				}
			} else {
				return errors.New("arg is not a keyer")
			}
		}
		return nil
	}

	switch req.cmd {
	case add:
		req.err <- loopArgs(func(key string, arg interface{}) error {
			// add or update it, here only myself refer the map
			if poster, ok := arg.(Poster); ok {
				d.data[key] = arg.(Poster)
				dprintf("Add: key(%s), title(%s), date(%s)\n",
					poster.Key(), poster.Title(), poster.Date())
			}
			return nil
		})
		return
	case remove:
		req.err <- loopArgs(func(key string, arg interface{}) error {
			if poster, ok := d.data[key]; ok {
				delete(d.data, key)
				dprintf("Remove: key(%s), title(%s), date(%s)\n",
					key, poster.Title(), poster.Date())
			}
			return nil
		})
		return
	case get:
		content := make([]Poster, 0)
		err := loopArgs(func(key string, arg interface{}) error {
			if v, found := d.data[key]; found {
				content = append(content, v)
				return nil
			}

			// not found
			return noFound
		})

		// some internal error
		if err != nil {
			req.err <- err
			return
		}

		// get all
		if len(content) == 0 {
			for _, v := range d.data {
				content = append(content, v)
			}
		}

		req.result <- content
		req.err <- nil
	}
}

type cmd int

const (
	add cmd = iota
	remove
	get
)

type request struct {
	cmd    cmd
	args   []interface{}
	result chan []Poster
	err    chan error
}

// Add add something into the dataCenter
// If the things are exist, update it
// Some internal error will be returned
func (s *Storage) Add(args ...Poster) error {
	r := &request{
		cmd:  add,
		args: make([]interface{}, len(args)),
		err:  make(chan error, 1),
	}
	for i, p := range args {
		r.args[i] = p
	}
	s.requestCh <- r
	return <-r.err
}

// Remove remove something from the dataCenter
// If the things are not exist, do nothing
// Some internal error will be returned
func (s *Storage) Remove(args ...Keyer) error {
	r := &request{
		cmd:  remove,
		args: make([]interface{}, len(args)),
		err:  make(chan error, 1),
	}
	for i, k := range args {
		r.args[i] = k
	}
	s.requestCh <- r
	return <-r.err
}

// Response for the request
type Result struct {
	Content []Poster
}

// Satisfy sort.Interface
func (r *Result) Len() int {
	return len(r.Content)
}

func (r *Result) Less(i, j int) bool {
	return r.Content[i].Date().After(r.Content[j].Date())
}

func (r *Result) Swap(i, j int) {
	r.Content[i], r.Content[j] = r.Content[j], r.Content[i]
}

// Get may get something from the dataCenter
// If you want get sth special, give the filter arg
// Otherwise, get all
// Some internal error will be returned
func (s *Storage) Get(args ...Keyer) (*Result, error) {
	r := &request{
		cmd:    get,
		args:   make([]interface{}, len(args)),
		result: make(chan []Poster, 1),
		err:    make(chan error, 1),
	}
	for i, k := range args {
		r.args[i] = k
	}
	s.requestCh <- r
	if err := <-r.err; err != nil {
		return nil, err
	}
	return &Result{<-r.result}, nil
}

// Destroys this storage
func (s *Storage) Destroy() {
	s.closeCh <- struct{}{}
}
