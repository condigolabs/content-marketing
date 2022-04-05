package services

import (
	"errors"
	"github.com/sirupsen/logrus"
	"io"
)

// Closers represents a slice of Closer and is a Closer itself.
type Closers []io.Closer

// Close calls Close function of each slice Closer entry.
func (c *Closers) Close() error {
	ret := MultiError{}
	if c != nil {
		for _, closer := range *c {
			if closer != nil {
				err := closer.Close()
				if err != nil {
					ret.Add(err)
				}
			}
		}
	}
	return ret.Err()
}

type CloserWrapper func()

func (c CloserWrapper) Close() error {
	c()
	return nil
}

type CloserErrWrapper func() error

func (c CloserErrWrapper) Close() error {
	return c()
}

func SafeClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		logrus.WithError(err).Error("Error while closing closer!")
	}
}

type MultiError []error

func (e *MultiError) Add(errors ...error) *MultiError {
	for _, err := range errors {
		if err != nil {
			*e = append(*e, err)
		}
	}
	return e
}

func (e *MultiError) Err() error {
	if e == nil || len(*e) == 0 {
		return nil
	}

	if len(*e) == 1 {
		return (*e)[0]
	}

	var message string
	for i, err := range *e {
		if i > 0 {
			message += "\n"
		}
		message += err.Error()
	}
	return errors.New(message)
}
