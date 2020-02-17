package storage

// Keyer represent a key to post
type Keyer interface {
	// key returns a strong for a key
	Key() string
}

// StringKey is literally string as a key
type StringKey string

func (s StringKey) Key() string { return string(s) }

var (
	_ Keyer = *(*StringKey)(nil)
)
