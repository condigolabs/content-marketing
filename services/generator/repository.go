package generator

import "io"

type Generator interface {
	io.Closer
}
