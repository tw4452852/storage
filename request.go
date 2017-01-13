package storage

type cmd int

const (
	ADD cmd = iota
	REMOVE
	GET
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
func Add(args ...Poster) error {
	r := &request{
		cmd:  ADD,
		args: make([]interface{}, len(args)),
		err:  make(chan error, 1),
	}
	for i, p := range args {
		r.args[i] = p
	}
	dataCenter.requestCh <- r
	return <-r.err
}

// Remove remove something from the dataCenter
// If the things are not exist, do nothing
// Some internal error will be returned
func Remove(args ...Keyer) error {
	r := &request{
		cmd:  REMOVE,
		args: make([]interface{}, len(args)),
		err:  make(chan error, 1),
	}
	for i, k := range args {
		r.args[i] = k
	}
	dataCenter.requestCh <- r
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
func Get(args ...Keyer) (*Result, error) {
	r := &request{
		cmd:    GET,
		args:   make([]interface{}, len(args)),
		result: make(chan []Poster, 1),
		err:    make(chan error, 1),
	}
	for i, k := range args {
		r.args[i] = k
	}
	dataCenter.requestCh <- r
	if err := <-r.err; err != nil {
		return nil, err
	}
	return &Result{<-r.result}, nil
}
