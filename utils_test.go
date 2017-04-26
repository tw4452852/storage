package storage

import (
	"testing"
)

func TestBytes2String(t *testing.T) {
	for name, c := range map[string]struct {
		input  []byte
		expect string
	}{
		"nil": {},
		"blank": {
			input: []byte{},
		},
		"normal": {
			input:  []byte{'t', 'w'},
			expect: "tw",
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			if got := Bytes2String(c.input); got != c.expect {
				t.Errorf("result not match: expect %v, got %v", c.expect, got)
			}
		})
	}
}
