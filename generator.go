package storage

import (
	"io"
)

type Generator interface {
	// Match return if the filename is match this Generator
	Match(filename string) bool
	// Generate the meta data of a post
	Generate(io.Reader, Staticer) (Poster, error)
}

var generators []Generator

func RegisterGenerator(gen Generator) {
	generators = append(generators, gen)
}

// FindGenerator find the first match Generator
func FindGenerator(filename string) Generator {
	for _, gen := range generators {
		if gen.Match(filename) {
			return gen
		}
	}
	return nil
}
