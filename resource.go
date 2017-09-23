package robospider

import "io"

type Resource struct {
	Name  string
	Found bool
	Body  io.ReadCloser
}
