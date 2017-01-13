package storage

import (
	"errors"
	"log"
)

type storage struct {
	requestCh chan *request     //for outcoming request
	closeCh   chan bool         //for exit
	data      map[string]Poster //internal data storage
}

var (
	dataCenter *storage
	debug      = true
)

func initStorage() {
	dataCenter = &storage{
		requestCh: make(chan *request),
		closeCh:   make(chan bool),
		data:      make(map[string]Poster),
	}
	go dataCenter.serve()
}

func (d *storage) serve() {
	for {
		select {
		case req := <-d.requestCh:
			d.handleRequest(req)
		case <-d.closeCh:
			return
		}
	}
}

func (d *storage) handleRequest(req *request) {
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
	case ADD:
		req.err <- loopArgs(func(key string, arg interface{}) error {
			// add or update it, here only myself refer the map
			if poster, ok := arg.(Poster); ok {
				d.data[key] = arg.(Poster)
				if debug {
					log.Printf("Add: key(%s), title(%s), date(%s)\n",
						poster.Key(), poster.Title(), poster.Date())
				}
			}
			return nil
		})
		return
	case REMOVE:
		req.err <- loopArgs(func(key string, arg interface{}) error {
			if poster, ok := d.data[key]; ok {
				delete(d.data, key)
				if debug {
					log.Printf("Remove: key(%s), title(%s), date(%s)\n",
						key, poster.Title(), poster.Date())
				}
			}
			return nil
		})
		return
	case GET:
		content := make([]Poster, 0)
		err := loopArgs(func(key string, arg interface{}) error {
			if v, found := d.data[key]; found {
				content = append(content, v)
				return nil
			}

			// not found
			return errors.New("can't find want you want")
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

func (d *storage) reset() {
	d.data = make(map[string]Poster)
}

func (d *storage) find(key string) interface{} {
	if v, found := d.data[key]; found {
		return v
	}
	return nil
}
